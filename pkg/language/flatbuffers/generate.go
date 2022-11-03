package flatbuffers

import (
	"fmt"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"log"
	"path"
	"sort"
	"strings"
)

func (*fbsLang) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var regularFbsFiles []string
	for _, name := range args.RegularFiles {
		if strings.HasSuffix(name, ".flatbuffers") {
			regularFbsFiles = append(regularFbsFiles, name)
		}
	}

	consumedFileSet := make(map[string]bool)
	for _, r := range args.OtherGen {
		if r.Kind() != "fbs_library" {
			continue
		}
		for _, f := range r.AttrStrings("srcs") {
			consumedFileSet[f] = true
		}
	}

	var genFbsFiles, genFbsFilesNotConsumed []string
	for _, name := range args.GenFiles {
		if strings.HasSuffix(name, ".flatbuffers") {
			genFbsFiles = append(genFbsFiles, name)
			if !consumedFileSet[name] {
				genFbsFilesNotConsumed = append(genFbsFilesNotConsumed, name)
			}
		}
	}
	pkgs := buildPackages(args.Config, args.Dir, args.Rel, regularFbsFiles)
	shouldSetVisibility := args.File == nil || !args.File.HasDefaultVisibility()
	var res language.GenerateResult
	for _, pkg := range pkgs {
		r := generateFbs(args.Config, args.Rel, pkg, shouldSetVisibility)
		if r.IsEmpty(fbsKinds[r.Kind()]) {
			res.Empty = append(res.Empty, r)
		} else {
			res.Gen = append(res.Gen, r)
		}
	}
	sort.SliceStable(res.Gen, func(i, j int) bool {
		return res.Gen[i].Name() < res.Gen[j].Name()
	})
	res.Imports = make([]interface{}, len(res.Gen))
	for i, r := range res.Gen {
		res.Imports[i] = r.PrivateAttr(config.GazelleImportsKey)
	}
	res.Empty = append(res.Empty, generateEmpty(args.File, regularFbsFiles)...)
	return res
}

func RuleName(names ...string) string {
	base := "root"
	for _, name := range names {
		notIdent := func(c rune) bool {
			return !('A' <= c && c <= 'Z' ||
				'a' <= c && c <= 'z' ||
				'0' <= c && c <= '9' ||
				c == '_')
		}
		if i := strings.LastIndexFunc(name, notIdent); i >= 0 {
			name = name[i+1:]
		}
		if name != "" {
			base = name
			break
		}
	}
	return base + "_fbs"
}

func buildPackages(c *config.Config, dir, rel string, fbsFiles []string) []*Package {
	packageMap := make(map[string]*Package)
	for _, name := range fbsFiles {
		info := fbsFileInfo(dir, name)
		key := info.PackageName

		if packageMap[key] == nil {
			packageMap[key] = newPackage(info.PackageName)
		}
		packageMap[key].addFile(info)
	}

	pkg, err := selectPackage(dir, rel, packageMap)
	if err != nil {
		log.Print(err)
	}
	if pkg == nil {
		return nil
	}
	return []*Package{pkg}
}

func selectPackage(dir, rel string, packageMap map[string]*Package) (*Package, error) {
	if len(packageMap) == 0 {
		return nil, nil
	}

	if len(packageMap) == 1 {
		for _, pkg := range packageMap {
			return pkg, nil
		}
	}
	defaultPackageName := strings.Replace(rel, "/", "_", -1)
	for _, pkg := range packageMap {
		if pkgName := goPackageName(pkg); pkgName != "" && pkgName == defaultPackageName {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("%s: directory contains multiple flatbuffers packages. Gazelle can only generate an fbs_library for one package.", dir)
}

func goPackageName(pkg *Package) string {
	if pkg.Name != "" {
		return strings.Replace(pkg.Name, ".", "_", -1)
	}
	if len(pkg.Files) == 1 {
		for s := range pkg.Files {
			return strings.TrimSuffix(s, ".flatbuffers")
		}
	}
	return ""
}

func generateFbs(c *config.Config, rel string, pkg *Package, shouldSetVisibility bool) *rule.Rule {
	var name string
	name = RuleName(goPackageName(pkg), rel)
	r := rule.NewRule("fbs_library", name)
	srcs := make([]string, 0, len(pkg.Files))
	for f := range pkg.Files {
		srcs = append(srcs, f)
	}
	sort.Strings(srcs)
	if len(srcs) > 0 {
		r.SetAttr("srcs", srcs)
	}
	r.SetPrivateAttr("_package", *pkg)
	imports := make([]string, 0, len(pkg.Imports))
	for i := range pkg.Imports {
		if _, ok := pkg.Files[path.Base(i)]; ok && getPrefix(c, path.Dir(i)) == getPrefix(c, rel) {
			delete(pkg.Imports, i)
			continue
		}
		imports = append(imports, i)
	}
	sort.Strings(imports)
	r.SetPrivateAttr(config.GazelleImportsKey, imports)
	for k, v := range pkg.Options {
		r.SetPrivateAttr(k, v)
	}
	if shouldSetVisibility {
		vis := rule.CheckInternalVisibility(rel, "//visibility:public")
		r.SetAttr("visibility", []string{vis})
	}
	return r
}

func getPrefix(c *config.Config, rel string) string {
	prefix := rel
	return prefix
}

func generateEmpty(f *rule.File, regularFiles []string) []*rule.Rule {
	if f == nil {
		return nil
	}
	knownFiles := make(map[string]bool)
	for _, f := range regularFiles {
		knownFiles[f] = true
	}
	var empty []*rule.Rule
outer:
	for _, r := range f.Rules {
		if r.Kind() != "fbs_library" {
			continue
		}
		srcs := r.AttrStrings("srcs")
		if len(srcs) == 0 && r.Attr("srcs") != nil {
			continue
		}
		for _, src := range srcs {
			if knownFiles[src] {
				continue outer
			}
		}
		empty = append(empty, rule.NewRule("fbs_library", r.Name()))
	}
	return empty
}
