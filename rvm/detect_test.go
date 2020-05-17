package rvm_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit"

	"github.com/avarteqgmbh/rvm-go-cnb/rvm"
	"github.com/avarteqgmbh/rvm-go-cnb/rvm/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		cnbDir     string
		workingDir string

		rubyVersionParser *fakes.VersionParser
		gemFileParser     *fakes.VersionParser
		gemFileLockParser *fakes.VersionParser
		detect            packit.DetectFunc
	)

	it.Before(func() {
		rubyVersionParser = &fakes.VersionParser{}
		gemFileParser = &fakes.VersionParser{}
		gemFileLockParser = &fakes.VersionParser{}

		logEmitter := rvm.NewLogEmitter(os.Stdout)
		detect = rvm.Detect(logEmitter, rubyVersionParser, gemFileParser, gemFileLockParser)
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

			err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte(`api = "0.2"
[buildpack]
	id = "org.some-org.some-buildpack"
	name = "Some Buildpack"
	version = "some-version"

	[metadata.configuration]
		uri = "https://get.rvm.io"
		default_rvm_version = "1.29.10"
		default_ruby_version = "2.7.1"
`), 0644)
			Expect(err).NotTo(HaveOccurred())

			workingDir, err = ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(workingDir, "Gemfile"), []byte(`source 'https://rubygems.org'`), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		it("returns a plan that provides RVM and requires node", func() {
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
					{
						Name:    "node",
						Version: rvm.DefaultNodeVersion,
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
			Expect(os.RemoveAll(layersDir)).To(Succeed())
			Expect(os.RemoveAll(cnbDir)).To(Succeed())
		})
	})
}
