package adapters

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// needsTool skips when a tool isn't on PATH (e.g. a restrictive sandbox), keeping these
// subprocess-backed adapter tests robust.
func needsTool(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not available: %v", name, err)
	}
}

func TestCaptureReturnsStdout(t *testing.T) {
	needsTool(t, "printf")
	out, err := Capture("printf", []string{"%s", "hello"}, ExecOptions{})
	require.NoError(t, err)
	assert.Equal(t, "hello", strings.TrimSpace(out))
}

func TestCaptureFeedsStdin(t *testing.T) {
	needsTool(t, "cat")
	out, err := Capture("cat", nil, ExecOptions{Input: "piped"})
	require.NoError(t, err)
	assert.Equal(t, "piped", strings.TrimSpace(out))
}

func TestCaptureRejectsNonZeroWithStderr(t *testing.T) {
	needsTool(t, "sh")
	_, err := Capture("sh", []string{"-c", "echo boom >&2; exit 3"}, ExecOptions{})
	assert.ErrorContains(t, err, "boom")
}

func TestRunRejectsNonZero(t *testing.T) {
	needsTool(t, "sh")
	assert.Error(t, Run("sh", []string{"-c", "exit 1"}, ExecOptions{}))
}

func TestRunResolvesOnSuccess(t *testing.T) {
	needsTool(t, "true")
	assert.NoError(t, Run("true", nil, ExecOptions{}))
}
