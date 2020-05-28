package rvm_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitRvm(t *testing.T) {
	suite := spec.New("rvm", spec.Report(report.Terminal{}))
	suite("Detect", testDetect)
	suite.Run(t)
}