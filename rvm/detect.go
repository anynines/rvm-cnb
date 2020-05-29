package rvm

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

// VersionParser represents a parser for files like .ruby-version and Gemfiles
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

// BuildPlanMetadata represents this buildpack's metadata
type BuildPlanMetadata struct {
	RubyVersion string `toml:"ruby_version"`
}

// NodebuildPlanMetadata represents the metadata for the node dependency
type NodebuildPlanMetadata struct {
	Build  bool `toml:"build"`
	Launch bool `toml:"launch"`
}

// VersionParserEnv represents an environment that contains everything that is
// needed to execute a particular ruby version parser
type VersionParserEnv struct {
	parser  VersionParser
	path    string
	context packit.DetectContext
	logger  LogEmitter
}

// ParseVersion is a generalized function that parses a particular ruby version
// source
func ParseVersion(env VersionParserEnv, version *string) error {
	fullPath := filepath.Join(env.context.WorkingDir, env.path)
	parseResultRubyVersion, err := env.parser.ParseVersion(fullPath)
	if err == nil && parseResultRubyVersion != "" {
		*version = parseResultRubyVersion
		env.logger.Detail("Found Ruby version in %s: %s", fullPath, *version)
		return nil
	}
	return err
}

// Detect whether this buildpack should install RVM
func Detect(logger LogEmitter, rubyVersionParser VersionParser, gemFileParser VersionParser, gemFileLockParser VersionParser) packit.DetectFunc {
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

		versionEnvs := []VersionParserEnv{
			{
				parser:  rubyVersionParser,
				path:    ".ruby-version",
				context: context,
				logger:  logger,
			},
			{
				parser:  gemFileParser,
				path:    "Gemfile",
				context: context,
				logger:  logger,
			},
			{
				parser:  gemFileLockParser,
				path:    "Gemfile.lock",
				context: context,
				logger:  logger,
			},
		}

		for _, env := range versionEnvs {
			err = ParseVersion(env, &rubyVersion)
			if err != nil {
				logger.Detail("Parsing '%s' failed", env.path)
				return packit.DetectResult{}, err
			}
		}

		logger.Detail("Detected Ruby version: %s", rubyVersion)
		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "rvm",
						Metadata: BuildPlanMetadata{
							RubyVersion: rubyVersion,
						},
					},
					{
						Name:    "node",
						Version: configuration.DefaultNodeVersion,
						Metadata: NodebuildPlanMetadata{
							Build:  true,
							Launch: true,
						},
					},
				},
			},
		}, nil
	}
}
