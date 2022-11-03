package flatbuffers

import (
	"errors"
	"fmt"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"log"
	"path"
	"sort"
	"strings"
)

func (*fbsLang) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	rel := f.Pkg
	srcs := r.AttrStrings("srcs")
	imports := make([]resolve.ImportSpec, len(srcs))
	prefix := rel
	for i, src := range srcs {
		imports[i] = resolve.ImportSpec{Lang: "flatbuffers", Imp: path.Join(prefix, src)}
	}
	return imports
}

func (*fbsLang) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

func (*fbsLang) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, importsRaw interface{}, from label.Label) {
	if importsRaw == nil {
		return
	}
	imports := importsRaw.([]string)
	r.DelAttr("deps")
	depSet := make(map[string]bool)
	for _, imp := range imports {
		l, err := resolveFbs(c, ix, r, imp, from)
		if err == errSkipImport {
			continue
		} else if err != nil {
			log.Print(err)
		} else {
			l = l.Rel(from.Repo, from.Pkg)
			depSet[l.String()] = true
		}
	}
	if len(depSet) > 0 {
		deps := make([]string, 0, len(depSet))
		for dep := range depSet {
			deps = append(deps, dep)
		}
		sort.Strings(deps)
		r.SetAttr("deps", deps)
	}
}

var (
	errSkipImport = errors.New("std import")
	errNotFound   = errors.New("not found")
)

func resolveFbs(c *config.Config, ix *resolve.RuleIndex, r *rule.Rule, imp string, from label.Label) (label.Label, error) {
	if !strings.HasSuffix(imp, ".flatbuffers") {
		return label.NoLabel, fmt.Errorf("can't import non-flatbuffers: %q", imp)
	}

	if l, ok := resolve.FindRuleWithOverride(c, resolve.ImportSpec{Imp: imp, Lang: "flatbuffers"}, "flatbuffers"); ok {
		return l, nil
	}

	if l, err := resolveWithIndex(c, ix, imp, from); err == nil || err == errSkipImport {
		return l, err
	} else if err != errNotFound {
		return label.NoLabel, err
	}

	rel := path.Dir(imp)
	if rel == "." {
		rel = ""
	}
	name := RuleName(rel)
	return label.New("", rel, name), nil
}

func resolveWithIndex(c *config.Config, ix *resolve.RuleIndex, imp string, from label.Label) (label.Label, error) {
	matches := ix.FindRulesByImportWithConfig(c, resolve.ImportSpec{Imp: imp, Lang: "flatbuffers"}, "flatbuffers")
	if len(matches) == 0 {
		return label.NoLabel, errNotFound
	}
	if len(matches) > 1 {
		return label.NoLabel, fmt.Errorf("multiple rules (%s and %s) may be imported with %q from %s", matches[0].Label, matches[1].Label, imp, from)
	}
	if matches[0].IsSelfImport(from) {
		return label.NoLabel, errSkipImport
	}
	return matches[0].Label, nil
}
