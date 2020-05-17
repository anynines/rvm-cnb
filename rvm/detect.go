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

		rvFile, err := os.Open(filepath.Join(context.WorkingDir, ".ruby-version"))
		if err == nil {
			defer rvFile.Close()
			bytes, err := ioutil.ReadAll(rvFile)
			if err == nil {
				rubyVersion = strings.TrimSpace(string(bytes))
				logger.Detail("Found Ruby version in .ruby-version: %s", rubyVersion)
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
			re, err := regexp.Compile(`^ruby ["']?([[:alnum:]\.\-]+)["']?`)
			if err == nil {
				for scanner.Scan() {
					rubySubSlices := re.FindSubmatch([]byte(scanner.Text()))
					if rubySubSlices != nil {
						rubyVersion = string(rubySubSlices[1])
						logger.Detail("Found Ruby version in Gemfile: %s", rubyVersion)
					}
				}
			}
		}

		gemFileLock, err := os.Open(filepath.Join(context.WorkingDir, "Gemfile.lock"))
		if err == nil {
			defer gemFileLock.Close()
			scanner := bufio.NewScanner(gemFileLock)
			for scanner.Scan() {
				if strings.TrimSpace(scanner.Text()) == "RUBY VERSION" {
					if scanner.Scan() {
						rubyVersion = strings.TrimSpace(strings.Trim(scanner.Text(), "ruby "))
						logger.Detail("Found Ruby version in Gemfile.lock: %s", rubyVersion)
						break
					}
				}
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
