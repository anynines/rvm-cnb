package rvm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
)

// Env represents an RVM environment
type Env struct {
	context       packit.BuildContext
	logger        LogEmitter
	configuration Configuration
	environment   EnvironmentConfiguration
}

// BuildRvm builds the RVM environment
func (r Env) BuildRvm() (packit.BuildResult, error) {
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

// RunBashCmd executes a command using BASH
func (r Env) RunBashCmd(command string, rvmLayer *packit.Layer) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = append(
		os.Environ(),
		DefaultVariables(rvmLayer)...,
	)

	r.logger.Process("Executing: %s", strings.Join(cmd.Args, " "))
	r.logger.Subprocess("Environment variables:\n%s", strings.Join(cmd.Env, "\n"))
	r.logger.Break()

	var stdOutBytes bytes.Buffer
	cmd.Stdout = &stdOutBytes

	var stdErrBytes bytes.Buffer
	cmd.Stderr = &stdErrBytes

	err := cmd.Run()

	if err != nil {
		r.logger.Subprocess("Command failed: %s", cmd.String())
		r.logger.Subprocess("Command stderr: %s", stdErrBytes)
		r.logger.Subprocess("Error status code: %s", err.Error())
		return err
	}

	r.logger.Subprocess("Command succeeded: %s", cmd.String())
	r.logger.Subprocess("Command output: %s", stdOutBytes)

	return nil
}

// RunRvmCmd executes a command in an RVM environment
func (r Env) RunRvmCmd(command string, rvmLayer *packit.Layer) error {
	profileDScript := filepath.Join(rvmLayer.Path, "profile.d", "rvm")
	fullRvmCommand := strings.Join([]string{
		"source",
		profileDScript,
		"&&",
		command,
	}, " ")

	return r.RunBashCmd(fullRvmCommand, rvmLayer)
}

func (r Env) rubyVersion() string {
	rubyVersion := r.configuration.DefaultRubyVersion
	for _, entry := range r.context.Plan.Entries {
		if entry.Name == "rvm" {
			rubyVersion = fmt.Sprintf("%v", entry.Metadata["ruby_version"])
		}
	}
	return rubyVersion
}

func (r Env) installRVM() (packit.BuildResult, error) {
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
	// cmd := exec.Command("bash", "-c", shellCmd)
	err = r.RunBashCmd(shellCmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	autolibsCmd := strings.Join([]string{
		filepath.Join(rvmLayer.Path, "bin", "rvm"),
		"autolibs",
		"0",
	}, " ")
	err = r.RunRvmCmd(autolibsCmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	rubyInstallCmd := strings.Join([]string{
		filepath.Join(rvmLayer.Path, "bin", "rvm"),
		"install",
		r.rubyVersion(),
	}, " ")
	err = r.RunRvmCmd(rubyInstallCmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	gemUpdateSystemCmd := strings.Join([]string{
		"gem",
		"update",
		"-N",
		"--system",
	}, " ")

	err = r.RunRvmCmd(gemUpdateSystemCmd, &rvmLayer)
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
