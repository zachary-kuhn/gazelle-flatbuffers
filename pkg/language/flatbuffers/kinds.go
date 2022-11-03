package flatbuffers

import "github.com/bazelbuild/bazel-gazelle/rule"

var fbsKinds = map[string]rule.KindInfo{
	"fbs_library": {
		MatchAttrs:    []string{"srcs"},
		NonEmptyAttrs: map[string]bool{"srcs": true},
		MergeableAttrs: map[string]bool{
			"srcs": true,
		},
		ResolveAttrs: map[string]bool{"deps": true},
	},
}

var fbsLoads = []rule.LoadInfo{
	{
		Name: "@com_github_google_flatbuffers:build_deps.bzl",
		Symbols: []string{
			"flatbuffer_library_public",
		},
	},
}

func (*fbsLang) Kinds() map[string]rule.KindInfo { return fbsKinds }
func (*fbsLang) Loads() []rule.LoadInfo          { return fbsLoads }
