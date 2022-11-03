package flatbuffers

import (
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/zachary-kuhn/gazelle-flatbuffers/pkg/language/flatbuffers"
)

func NewLanguage() language.Language {
	return flatbuffers.NewFlatbuffersLanguage()
}
