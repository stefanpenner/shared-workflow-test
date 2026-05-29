package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	assert.Equal(t, "▸ Lint\n  paths   src\n  config  .x", Report("src", ".x"))
}

func TestReportDefaults(t *testing.T) {
	got := Report("  ", "")
	assert.Contains(t, got, "paths   .")
	assert.Contains(t, got, "config  .eslintrc")
}
