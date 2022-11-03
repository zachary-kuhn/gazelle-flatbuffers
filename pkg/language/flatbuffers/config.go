package flatbuffers

import (
	"flag"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func (*fbsLang) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {

}

func (*fbsLang) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	return nil
}

func (*fbsLang) KnownDirectives() []string {
	return []string{}
}

func (*fbsLang) Configure(c *config.Config, rel string, f *rule.File) {

}
