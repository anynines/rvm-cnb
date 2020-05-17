package rvm

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/packit"
)

const (
	// DefaultNodeVersion specifies the default NodeJS version to be installed by
	// the Node CNB.
	DefaultNodeVersion string = "12.*"
)

type buildPlanMetadata struct {
	RubyVersion string `toml:"ruby_version"`
}

type nodebuildPlanMetadata struct {
	Build  bool `toml:"build"`
	Launch bool `toml:"launch"`
}

// Detect whether this buildpack should install RVM
func Detect(logger LogEmitter) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, err := os.Stat(filepath.Join(context.WorkingDir, "Gemfile"))
		if os.IsNotExist(err) {
			return packit.DetectResult{}, err
		}

		configuration, err := ReadConfiguration(context.CNBPath)
		if err != nil {
			return packit.DetectResult{}, err
		}

		rubyVersion := configuration.DefaultRubyVersion

		rubyVersionParser := NewRubyVersionParser()
		rubyVersionPath := filepath.Join(context.WorkingDir, ".ruby-version")
		parseResultRubyVersion, err := rubyVersionParser.ParseVersion(rubyVersionPath)
		if err == nil && parseResultRubyVersion != "" {
			rubyVersion = parseResultRubyVersion
			logger.Detail("Found Ruby version in %s: %s", rubyVersionPath, rubyVersion)
		}

		gemFileParser := NewGemfileParser()
		gemFilePath := filepath.Join(context.WorkingDir, "Gemfile")
		parseResultGemfile, err := gemFileParser.ParseVersion(gemFilePath)
		if err == nil && parseResultGemfile != "" {
			rubyVersion = parseResultGemfile
			logger.Detail("Found Ruby version in %s: %s", gemFilePath, rubyVersion)
		}

		gemFileLockParser := NewGemfileLockParser()
		gemFileLockPath := filepath.Join(context.WorkingDir, "Gemfile.lock")
		parseResultGemfileLock, err := gemFileLockParser.ParseVersion(gemFileLockPath)
		if err == nil && parseResultGemfileLock != "" {
			rubyVersion = parseResultGemfileLock
			logger.Detail("Found Ruby version in %s: %s", gemFileLockPath, rubyVersion)
		}

		logger.Detail("Detected Ruby version: %s", rubyVersion)
		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name:     "rvm",
						Metadata: buildPlanMetadata{RubyVersion: rubyVersion},
					},
					{
						Name:    "node",
						Version: DefaultNodeVersion,
						Metadata: nodebuildPlanMetadata{
							Build:  true,
							Launch: true,
						},
					},
				},
			},
		}, nil
	}
}
