package main

import (
	"os"

	"github.com/avarteqgmbh/rvm-go-cnb/rvm"

	"github.com/cloudfoundry/packit"
)

func main() {
	logEmitter := rvm.NewLogEmitter(os.Stdout)
	rubyVersionParser := rvm.NewRubyVersionParser()
	gemFileParser := rvm.NewGemfileParser()
	gemFileLockParser := rvm.NewGemfileLockParser()
	packit.Detect(rvm.Detect(logEmitter, rubyVersionParser, gemFileParser, gemFileLockParser))
}
