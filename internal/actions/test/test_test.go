package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	assert.Equal(t, "▸ Test\n  suite     integration\n  coverage  enabled", Report("integration", "true"))
}

func TestReportDefaultsAndCoverageOff(t *testing.T) {
	got := Report("", "false")
	assert.Contains(t, got, "suite     unit")
	assert.Contains(t, got, "coverage  disabled")
}

func TestReportCoverageEnabledForNonFalse(t *testing.T) {
	assert.Contains(t, Report("unit", "yes"), "coverage  enabled")
}
