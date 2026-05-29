package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repoWithCommit builds a git repo with one deterministic commit and returns its HEAD SHA.
func repoWithCommit(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, exec.Command("git", "init", "-q", dir).Run())
	require.NoError(t, exec.Command("git", "-C", dir, "config", "commit.gpgsign", "false").Run())
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte(content), 0o644))
	require.NoError(t, ConfigureBotIdentity(dir))
	require.NoError(t, CommitAll(dir, "shadow: fixed message"))
	sha, err := HeadSha(dir)
	require.NoError(t, err)
	return sha
}

func TestCommitDeterminismSameInputsSameSHA(t *testing.T) {
	needsTool(t, "git")
	a := repoWithCommit(t, "same")
	b := repoWithCommit(t, "same")
	assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{40}$`), a)
	assert.Equal(t, a, b)
}

func TestCommitDifferentContentDifferentSHA(t *testing.T) {
	needsTool(t, "git")
	assert.NotEqual(t, repoWithCommit(t, "one"), repoWithCommit(t, "two"))
}
