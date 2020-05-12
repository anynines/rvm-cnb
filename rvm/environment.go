package rvm

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/packit"
)

// Environment represents a shell environment
type Environment struct {
	logger LogEmitter
}

// NewEnvironment returns a new Environment with a logger
func NewEnvironment(logger LogEmitter) Environment {
	return Environment{
		logger: logger,
	}
}

// Configure a shell environment for use with RVM
func (e Environment) Configure(env packit.Environment, path string) error {
	scriptsPath := filepath.Join(path, "scripts")
	env.Override("rvm_path", path)
	env.Override("rvm_scripts_path", scriptsPath)
	env.Override("rvm_autoupdate_flag", "0")

	profileDPath := filepath.Join(path, "profile.d")
	err := os.MkdirAll(profileDPath, os.ModePerm)
	if err != nil {
		e.logger.Detail("Creating directory '%s' failed", profileDPath)
		return err
	}

	err = os.Symlink(filepath.Join(scriptsPath, "rvm"), filepath.Join(profileDPath, "rvm"))
	if err != nil {
		e.logger.Detail("Creating symlink from '%s' to '%s' failed", filepath.Join(scriptsPath, "rvm"), filepath.Join(profileDPath, "rvm"))
		return err
	}

	return nil
}