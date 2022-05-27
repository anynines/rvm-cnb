package main

import (
	"os"

	"github.com/avarteqgmbh/rvm-cnb/rvm"

	"github.com/paketo-buildpacks/packit/v2"
)

func main() {
	logEmitter := rvm.NewLogEmitter(os.Stdout)
	environment := rvm.NewEnvironment(logEmitter)
	packit.Build(rvm.Build(environment, logEmitter))
}
