package debug

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
)

// echoExec returns the command line it was asked to run, so tests can assert which probes ran.
func echoExec(file string, args []string) (string, error) {
	return file + " " + strings.Join(args, " "), nil
}

func TestTreeReportWrapsAndUsesEnv(t *testing.T) {
	out, err := TreeReport(echoExec, map[string]string{"HOME": "/h", "GITHUB_WORKSPACE": "/w"})
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(out, "::group::Environment\n"))
	assert.True(t, strings.HasSuffix(out, "\n::endgroup::"))
	assert.Contains(t, out, "ls -la /h")
	assert.Contains(t, out, "ls -la /h/work/")
	assert.Contains(t, out, "find /w")
}

func TestTreeReportFallsBackOnProbeFailure(t *testing.T) {
	out, err := TreeReport(func(string, []string) (string, error) { return "", ErrProbe }, nil)
	require.NoError(t, err)
	assert.Contains(t, out, "(unavailable)")
	assert.Contains(t, out, "(not found)")
}

func TestTreeReportPropagatesUnexpectedError(t *testing.T) {
	boom := errors.New("boom")
	_, err := TreeReport(func(string, []string) (string, error) { return "", boom }, nil)
	assert.ErrorIs(t, err, boom)
}

func TestGitReportNoRepo(t *testing.T) {
	out, err := GitReport(func(string, []string) (string, error) { return "", ErrProbe })
	require.NoError(t, err)
	assert.Equal(t, ghactions.Group("Git status", "No git repository in working directory"), out)
}

func TestGitReportRendersStatusAndDiffs(t *testing.T) {
	out, err := GitReport(func(_ string, args []string) (string, error) { return strings.Join(args, " "), nil })
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(out, "::group::Git status\n"))
	assert.Contains(t, out, "unstaged changes:")
	assert.Contains(t, out, "staged changes:")
}
