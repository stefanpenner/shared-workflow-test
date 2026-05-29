package ghactions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderOutputs(t *testing.T) {
	assert.Equal(t, "node_version=20\n", RenderOutputs([]Pair{{"node_version", "20"}}))
	assert.Equal(t, "a=1\nb=2\n", RenderOutputs([]Pair{{"a", "1"}, {"b", "2"}}))
}

func TestAppendOutputWritesPairs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out")
	require.NoError(t, AppendOutput(path, []Pair{{"pr", "7"}, {"sha", "abc"}}))
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "pr=7\nsha=abc\n", string(b))
}

func TestAppendOutputErrorsWhenUnset(t *testing.T) {
	assert.Error(t, AppendOutput("", []Pair{{"x", "1"}}))
}

func TestAppendFileAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "summary")
	require.NoError(t, AppendFile(path, "## one\n"))
	require.NoError(t, AppendFile(path, "## two\n"))
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "## one\n## two\n", string(b))
}

func TestAppendFileErrorsWhenEmptyPath(t *testing.T) {
	assert.Error(t, AppendFile("", "x"))
}
