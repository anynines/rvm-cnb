package rvm

import (
	"github.com/cloudfoundry/packit"
)

// EnvironmentConfiguration represents an environment and a path to the RVM
// layer
type EnvironmentConfiguration interface {
	Configure(env packit.Environment, path string) error
}

// Build the RVM layer provided by this buildpack
func Build(environment EnvironmentConfiguration, logger LogEmitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		configuration, err := ReadConfiguration(context.CNBPath)
		if err != nil {
			return packit.BuildResult{}, err
		}

		rvmEnv := RvmEnv{
			configuration: configuration,
			context:       context,
			environment:   environment,
			logger:        logger,
		}

		return rvmEnv.BuildRvm()
	}
}
