package rvm_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/avarteqgmbh/rvm-cnb/rvm"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEnvironment(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cnbDir            string
		layersDir         string
		buildPackTomlPath string = "../test/fixtures/before/some_buildpack.toml"
		rvmEnv            rvm.Env
	)

	context("Environment configuration", func() {
		it.Before(func() {
			var err error
			cnbDir, err = ioutil.TempDir("", "cnb")
			Expect(err).NotTo(HaveOccurred())

			someBuildPackTomlFile, err := ioutil.ReadFile(buildPackTomlPath)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
			Expect(err).NotTo(HaveOccurred())

			layersDir, err = ioutil.TempDir("", "layers")
			Expect(err).NotTo(HaveOccurred())

			err = os.Mkdir(filepath.Join(layersDir, "scripts"), 0755)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(layersDir, "scripts", "rvm"), []byte{}, 0644)
			Expect(err).NotTo(HaveOccurred())

			configuration, err := rvm.ReadConfiguration(cnbDir)
			buildContext := packit.BuildContext{CNBPath: cnbDir}
			logEmitter := rvm.NewLogEmitter(os.Stdout)
			environment := rvm.NewEnvironment(logEmitter)
			rvmEnv = rvm.Env{
				Context:       buildContext,
				Logger:        logEmitter,
				Configuration: configuration,
				Environment:   environment,
			}
		})

		it("returns a list of default environment variables", func() {
			rvmLayer := packit.Layer{
				Name: "rvm",
				Path: layersDir,
			}
			defaultVariables := rvm.DefaultVariables(&rvmLayer)
			Expect(defaultVariables).To(Equal([]string{
				"rvm_path=" + rvmLayer.Path,
				"rvm_scripts_path=" + filepath.Join(rvmLayer.Path, "scripts"),
				"rvm_autoupdate_flag=0",
			}))
		})

		it("configures the environment variables for this CNB", func() {
			env := packit.Environment{}
			err := rvmEnv.Environment.Configure(env, layersDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(filepath.Join(layersDir, "profile.d", "rvm")).To(BeARegularFile())
			Expect(env).To(Equal(packit.Environment{
				"rvm_autoupdate_flag.override": "0",
				"rvm_path.override":            layersDir,
				"rvm_scripts_path.override":    filepath.Join(layersDir, "scripts"),
			}))
		})

		it.After(func() {
			Expect(os.RemoveAll(cnbDir)).To(Succeed())
			Expect(os.RemoveAll(layersDir)).To(Succeed())
		})
	})
}
