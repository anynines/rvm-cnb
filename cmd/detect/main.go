package main

import (
	"os"

	"github.com/avarteqgmbh/rvm-cnb/rvm"

	"github.com/paketo-buildpacks/packit"
)

func main() {
	logEmitter := rvm.NewLogEmitter(os.Stdout)
	rubyVersionParser := rvm.NewRubyVersionParser()
	gemFileParser := rvm.NewGemfileParser()
	gemFileLockParser := rvm.NewGemfileLockParser()
	bundlerVersionParser := rvm.NewBundlerVersionParser()
	packit.Detect(rvm.Detect(logEmitter, rubyVersionParser, gemFileParser, gemFileLockParser, bundlerVersionParser))
}
