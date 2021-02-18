package rvm_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitRvm(t *testing.T) {
	suite := spec.New("rvm", spec.Report(report.Terminal{}))
	suite("Configuration", testConfiguration)
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("Environment", testEnvironment)
	suite("GemFileParser", testGemFileParser)
	suite("GemFileLockParser", testGemFileLockParser)
	suite("RubyVersionParser", testRubyVersionParser)
	suite("Detect", testDetect)
	suite.Run(t)
}
