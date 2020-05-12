package rvm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/packit"
)

// RvmEnv represents an RVM environment
type RvmEnv struct {
	context       packit.BuildContext
	logger        LogEmitter
	configuration Configuration
	environment   EnvironmentConfiguration
}

// BuildRvm builds the RVM environment
func (r RvmEnv) BuildRvm() (packit.BuildResult, error) {
	r.logger.Title("%s %s", r.context.BuildpackInfo.Name, r.context.BuildpackInfo.Version)

	r.logger.Process("Using RVM URI: %s\n", r.configuration.URI)
	r.logger.Process("default RVM version: %s\n", r.configuration.DefaultRVMVersion)
	r.logger.Process("build plan Ruby version: %s\n", r.rubyVersion())

	buildResult, err := r.installRVM()
	if err != nil {
		return packit.BuildResult{}, err
	}
	return buildResult, nil
}

func (r RvmEnv) rubyVersion() string {
	rubyVersion := r.configuration.DefaultRubyVersion
	for _, entry := range r.context.Plan.Entries {
		if entry.Name == "rvm" {
			rubyVersion = fmt.Sprintf("%v", entry.Metadata["ruby_version"])
		}
	}
	return rubyVersion
}

func (r RvmEnv) installRVM() (packit.BuildResult, error) {
	rvmLayer, err := r.context.Layers.Get("rvm", packit.LaunchLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	if rvmLayer.Metadata["rvm_version"] != nil &&
		rvmLayer.Metadata["rvm_version"].(string) == r.configuration.DefaultRVMVersion {
		r.logger.Process("Reusing cached layer %s", rvmLayer.Path)
		return packit.BuildResult{
			Plan: r.context.Plan,
			Layers: []packit.Layer{
				rvmLayer,
			},
		}, nil
	}

	r.logger.Process("Installing RVM version '%s' from URI '%s'", r.configuration.DefaultRVMVersion, r.configuration.URI)

	if err = rvmLayer.Reset(); err != nil {
		r.logger.Process("Resetting RVM layer failed")
		return packit.BuildResult{}, err
	}

	rvmLayer.Metadata = map[string]interface{}{
		"rvm_version":  r.configuration.DefaultRVMVersion,
		"ruby_version": r.rubyVersion(),
	}

	rvmLayer.Build = true
	rvmLayer.Cache = true
	rvmLayer.Launch = true

	err = r.environment.Configure(rvmLayer.SharedEnv, rvmLayer.Path)
	if err != nil {
		return packit.BuildResult{}, err
	}

	shellCmd := strings.Join([]string{
		"curl",
		"-vsSL",
		r.configuration.URI,
		"| bash -s -- --version",
		r.configuration.DefaultRVMVersion,
	}, " ")
	cmd := exec.Command("bash", "-c", shellCmd)
	err = r.runCommand(cmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	cmd = exec.Command(filepath.Join(rvmLayer.Path, "bin", "rvm"), "autolibs", "0")
	err = r.runCommand(cmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	cmd = exec.Command(filepath.Join(rvmLayer.Path, "bin", "rvm"), "install", r.rubyVersion())
	err = r.runCommand(cmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	return packit.BuildResult{
		Plan: r.context.Plan,
		Layers: []packit.Layer{
			rvmLayer,
		},
	}, nil
}

func (r RvmEnv) runCommand(cmd *exec.Cmd, rvmLayer *packit.Layer) error {
	cmd.Env = append(
		os.Environ(),
		"rvm_path="+rvmLayer.Path,
		"rvm_scripts_path="+filepath.Join(rvmLayer.Path, "scripts"),
		"rvm_autoupdate_flag=0",
	)
	r.logger.Action("Executing: %s", strings.Join(cmd.Args, " "))
	r.logger.Subdetail("Environment variables:\n%s", strings.Join(cmd.Env, "\n"))
	r.logger.Break()
	err := cmd.Run()
	if err != nil {
		r.logger.Detail("Command failed: %s", cmd.String())
		r.logger.Detail("Error status code: %s", err.Error())
		return err
	}

	return nil
}
