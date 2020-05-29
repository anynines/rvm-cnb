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

func testRubyVersionParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workDir            string
		rubyVersionkParser rvm.RubyVersionParser
	)

	context("when a Gemfile.lock is present", func() {
		it.Before(func() {
			var err error

			workDir, err = ioutil.TempDir("", "workDir")
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(workDir, ".ruby-version"), []byte("2.6.5\n"), 0644)
			Expect(err).NotTo(HaveOccurred())

			rubyVersionkParser = rvm.NewRubyVersionParser()
		})

		it("returns the ruby version after parsing Gemfile.lock", func() {
			rubyVersion, err := rubyVersionkParser.ParseVersion(filepath.Join(workDir, ".ruby-version"))
			Expect(err).NotTo(HaveOccurred())
			Expect(rubyVersion).To(Equal("2.6.5"))
		})

		it.After(func() {
			Expect(os.RemoveAll(workDir)).To(Succeed())
		})
	})
}
