package rvm_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/avarteqgmbh/rvm-cnb/rvm"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testGemFileLockParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workDir           string
		gemFileLockParser rvm.GemfileLockParser
	)

	context("when a Gemfile.lock is present", func() {
		it.Before(func() {
			var err error

			workDir, err = ioutil.TempDir("", "workDir")
			Expect(err).NotTo(HaveOccurred())

			gemFileLock, err := ioutil.ReadFile("../test/fixtures/read_version_gemfile/Gemfile.lock")
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(workDir, "Gemfile.lock"), gemFileLock, 0644)
			Expect(err).NotTo(HaveOccurred())

			gemFileLockParser = rvm.NewGemfileLockParser()
		})

		it("returns the ruby version after parsing Gemfile.lock", func() {
			rubyVersion, err := gemFileLockParser.ParseVersion(filepath.Join(workDir, "Gemfile.lock"))
			Expect(err).NotTo(HaveOccurred())
			Expect(rubyVersion).To(Equal("2.6.5p114"))
		})

		it.After(func() {
			Expect(os.RemoveAll(workDir)).To(Succeed())
		})
	})
}
