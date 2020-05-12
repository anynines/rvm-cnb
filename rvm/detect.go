package rvm

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudfoundry/libcfbuildpack/helper"
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

		gemFileLock, err := os.Open(filepath.Join(context.WorkingDir, "Gemfile.lock"))
		if err == nil {
			defer gemFileLock.Close()
			bundledWithFound := false
			scanner := bufio.NewScanner(gemFileLock)
			for scanner.Scan() {
				if bundledWithFound {
					rubyVersion = strings.TrimSpace(scanner.Text())
					logger.Detail("Found Ruby version in Gemfile.lock: %s", rubyVersion)
					break
				}
				if scanner.Text() == "RUBY VERSION" {
					bundledWithFound = true
				}
			}
		}

		gemfile := filepath.Join(context.WorkingDir, "Gemfile")
		if exists, _ := helper.FileExists(gemfile); exists {
			file, err := os.Open(gemfile)
			if err != nil {
				return packit.DetectResult{}, err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			re, err := regexp.Compile(`^ruby "[[:alnum:]\.\-]"`)
			if err == nil {
				for scanner.Scan() {
					rubyVersionText := re.Find([]byte(scanner.Text()))
					if rubyVersionText != nil {
						rubyVersion = string(rubyVersionText)
						logger.Detail("Found Ruby version in Gemfile %s", rubyVersion)
					}
				}
			}
		}

		rvFile, err := os.Open(filepath.Join(context.WorkingDir, ".ruby-version"))
		if err == nil {
			defer rvFile.Close()
			bytes, err := ioutil.ReadAll(rvFile)
			if err == nil {
				rubyVersion = string(bytes)
				logger.Detail("Found Ruby version in .ruby-version: %s", rubyVersion)
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
