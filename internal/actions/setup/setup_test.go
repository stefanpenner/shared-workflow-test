package setup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveNodeVersionTrims(t *testing.T) {
	v, err := ResolveNodeVersion(" 20 ")
	require.NoError(t, err)
	assert.Equal(t, "20", v)
}

func TestResolveNodeVersionErrorsWhenEmpty(t *testing.T) {
	_, err := ResolveNodeVersion("   ")
	assert.EqualError(t, err, "node-version is required")
}

func TestReport(t *testing.T) {
	assert.Equal(t, "▸ Setup\n  project       demo\n  node version  20", Report("demo", "20"))
}

func TestReportFallsBackForEmptyProject(t *testing.T) {
	assert.Contains(t, Report("", "20"), "project       (unknown project)")
}
