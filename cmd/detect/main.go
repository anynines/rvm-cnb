package main

import (
	"os"

	"github.com/avarteqgmbh/rvm-go-cnb/rvm"

	"github.com/cloudfoundry/packit"
)

func main() {
	logEmitter := rvm.NewLogEmitter(os.Stdout)
	packit.Detect(rvm.Detect(logEmitter))
}
