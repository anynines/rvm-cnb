package rvm_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit"

	"github.com/avarteqgmbh/rvm-cnb/rvm"
	"github.com/avarteqgmbh/rvm-cnb/rvm/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		cnbDir     string
		workingDir string

		rubyVersionParser  *fakes.VersionParser
		gemFileParser      *fakes.VersionParser
		gemFileLockParser  *fakes.VersionParser
		buildpackYMLParser *fakes.VersionParser
		detect             packit.DetectFunc
	)

	it.Before(func() {
		rubyVersionParser = &fakes.VersionParser{}
		gemFileParser = &fakes.VersionParser{}
		gemFileLockParser = &fakes.VersionParser{}
		buildpackYMLParser = &fakes.VersionParser{}

		logEmitter := rvm.NewLogEmitter(os.Stdout)
		detect = rvm.Detect(logEmitter, rubyVersionParser, gemFileParser, gemFileLockParser, buildpackYMLParser)
	})

	it("returns a plan that does not provide rvm because no Gemfile was found", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).To(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{Provides: nil, Requires: nil, Or: nil}))
	})

	context("when the app presents a Gemfile", func() {
		it.Before(func() {
			var err error
			layersDir, err = ioutil.TempDir("", "layers")
			Expect(err).NotTo(HaveOccurred())

			cnbDir, err = ioutil.TempDir("", "cnb")
			Expect(err).NotTo(HaveOccurred())

			someBuildPackTomlFile, err := ioutil.ReadFile("../test/fixtures/before/some_buildpack.toml")
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
			Expect(err).NotTo(HaveOccurred())

			workingDir, err = ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			basicGemfile, err := ioutil.ReadFile("../test/fixtures/before/Gemfile")
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(workingDir, "Gemfile"), basicGemfile, 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		it("returns a plan that provides RVM and requires rvm", func() {
			result, err := detect(packit.DetectContext{
				CNBPath:    cnbDir,
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "rvm",
						Metadata: rvm.BuildPlanMetadata{
							RubyVersion: "2.7.1",
						},
					},
				},
			}))
		})

		it("returns a plan that provides RVM and determines the ruby version by reading .ruby-version", func() {
			rubyVersionPath := filepath.Join(workingDir, ".ruby-version")
			err := ioutil.WriteFile(rubyVersionPath, []byte("2.3.8\n"), 0644)
			Expect(err).NotTo(HaveOccurred())

			rubyVersionParser.ParseVersionCall.Receives.Path = rubyVersionPath
			rubyVersionParser.ParseVersionCall.Returns.Version = "2.3.8"

			result, err := detect(packit.DetectContext{
				CNBPath:    cnbDir,
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "rvm",
						Metadata: rvm.BuildPlanMetadata{
							RubyVersion: "2.3.8",
						},
					},
				},
			}))
		})

		it("returns a plan that provides RVM and determines the ruby version by reading the Gemfile", func() {
			rubyVersionGemfile, err := ioutil.ReadFile("../test/fixtures/read_version_gemfile/Gemfile")
			Expect(err).NotTo(HaveOccurred())

			gemFilePath := filepath.Join(workingDir, "Gemfile")
			err = ioutil.WriteFile(gemFilePath, rubyVersionGemfile, 0644)
			Expect(err).NotTo(HaveOccurred())

			rubyVersionParser.ParseVersionCall.Receives.Path = gemFilePath
			rubyVersionParser.ParseVersionCall.Returns.Version = "2.5.3"

			result, err := detect(packit.DetectContext{
				CNBPath:    cnbDir,
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "rvm",
						Metadata: rvm.BuildPlanMetadata{
							RubyVersion: "2.5.3",
						},
					},
				},
			}))
		})

		it("returns a plan that provides RVM and determines the ruby version by reading Gemfile.lock", func() {
			rubyVersionGemfileLock, err := ioutil.ReadFile("../test/fixtures/read_version_gemfile/Gemfile.lock")
			Expect(err).NotTo(HaveOccurred())

			gemFileLockPath := filepath.Join(workingDir, "Gemfile.lock")
			err = ioutil.WriteFile(gemFileLockPath, rubyVersionGemfileLock, 0644)
			Expect(err).NotTo(HaveOccurred())

			rubyVersionParser.ParseVersionCall.Receives.Path = gemFileLockPath
			rubyVersionParser.ParseVersionCall.Returns.Version = "2.5.3"

			result, err := detect(packit.DetectContext{
				CNBPath:    cnbDir,
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "rvm",
						Metadata: rvm.BuildPlanMetadata{
							RubyVersion: "2.5.3",
						},
					},
				},
			}))
		})

		it("returns a plan that provides RVM and determines the ruby version by reading buildpack.yml", func() {
			buildPackYML, err := ioutil.ReadFile("../test/fixtures/read_version_buildpack_yml/buildpack.yml")
			Expect(err).NotTo(HaveOccurred())

			buildPackYMLPath := filepath.Join(workingDir, "buildpack.yml")
			err = ioutil.WriteFile(buildPackYMLPath, buildPackYML, 0644)
			Expect(err).NotTo(HaveOccurred())

			rubyVersionParser.ParseVersionCall.Receives.Path = buildPackYMLPath
			rubyVersionParser.ParseVersionCall.Returns.Version = "2.5.3"

			result, err := detect(packit.DetectContext{
				CNBPath:    cnbDir,
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "rvm",
						Metadata: rvm.BuildPlanMetadata{
							RubyVersion: "2.5.3",
						},
					},
				},
			}))
		})

		it("returns a plan that provides RVM and requires node", func() {
			buildPackYML, err := ioutil.ReadFile("../test/fixtures/read_version_buildpack_yml/buildpack_require_node.yml")
			Expect(err).NotTo(HaveOccurred())

			buildPackYMLPath := filepath.Join(workingDir, "buildpack.yml")
			err = ioutil.WriteFile(buildPackYMLPath, buildPackYML, 0644)
			Expect(err).NotTo(HaveOccurred())

			rubyVersionParser.ParseVersionCall.Receives.Path = buildPackYMLPath
			rubyVersionParser.ParseVersionCall.Returns.Version = "2.5.3"

			buildPackYMLParsed, err := rvm.BuildpackYMLParse(buildPackYMLPath)
			Expect(err).NotTo(HaveOccurred())

			result, err := detect(packit.DetectContext{
				CNBPath:    cnbDir,
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "rvm",
						Metadata: rvm.BuildPlanMetadata{
							RubyVersion: "2.5.3",
						},
					},
					{
						Name:    "node",
						Version: buildPackYMLParsed.NodeVersion,
						Metadata: rvm.NodebuildPlanMetadata{
							Build:  true,
							Launch: true,
						},
					},
				},
			}))
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
			Expect(os.RemoveAll(cnbDir)).To(Succeed())
			Expect(os.RemoveAll(layersDir)).To(Succeed())
		})
	})
}
