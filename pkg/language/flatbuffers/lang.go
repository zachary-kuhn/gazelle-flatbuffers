package flatbuffers

import "github.com/bazelbuild/bazel-gazelle/language"

const fbsName = "flatbuffers"

type fbsLang struct{}

func (*fbsLang) Name() string { return fbsName }

func NewFlatbuffersLanguage() language.Language {
	return &fbsLang{}
}
