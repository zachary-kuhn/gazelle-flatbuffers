package flatbuffers

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type FileInfo struct {
	Path, Name  string
	PackageName string
	Imports     []string
	HasServices bool
}

var fbsRe = buildFbsRegexp()

func fbsFileInfo(dir, name string) FileInfo {
	info := FileInfo{
		Path: filepath.Join(dir, name),
		Name: name,
	}
	content, err := os.ReadFile(info.Path)
	if err != nil {
		log.Printf("%s: error reading flatbuffers file: %v", info.Path, err)
		return info
	}

	for _, match := range fbsRe.FindAllSubmatch(content, -1) {
		switch {
		case match[includeSubexpIndex] != nil:
			imp := string(match[includeSubexpIndex])
			imp = imp[1 : len(imp)-1]
			info.Imports = append(info.Imports, imp)
		}
	}
	sort.Strings(info.Imports)

	return info
}

const (
	includeSubexpIndex = 1
)

func buildFbsRegexp() *regexp.Regexp {
	includeStmt := `\binclude\s*"(?P<import>.*)""\s*;`
	fbsReSrc := strings.Join([]string{includeStmt}, "|")
	return regexp.MustCompile(fbsReSrc)
}
