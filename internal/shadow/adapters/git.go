package adapters

import (
	"fmt"
	"os"
	"strings"
)

func gitCapture(args []string, dir string) (string, error) {
	return Capture("git", args, ExecOptions{Dir: dir})
}

// CloneShallow shallow-clones a repo at a ref into dir, authenticated with a token.
func CloneShallow(repo, ref, dir, token string) error {
	url := fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", token, repo)
	return Run("git", []string{"clone", "--depth=1", "--branch", ref, url, dir}, ExecOptions{})
}

// ConfigureBotIdentity sets a committer identity for the shadow bot commit.
func ConfigureBotIdentity(dir string) error {
	if _, err := gitCapture([]string{"config", "user.name", "shadow-testing[bot]"}, dir); err != nil {
		return err
	}
	_, err := gitCapture([]string{"config", "user.email", "shadow-testing@users.noreply.github.com"}, dir)
	return err
}

// ResetBranchToEmptyTree starts the shadow branch from HEAD and clears the tracked tree, so the
// next commit contains only the mirrored consumer files.
func ResetBranchToEmptyTree(dir, branch string) error {
	if _, err := gitCapture([]string{"checkout", "-B", branch}, dir); err != nil {
		return err
	}
	_, err := gitCapture([]string{"rm", "-rf", "--quiet", "."}, dir)
	return err
}

// deterministicDate fixes the commit timestamp so identical (tree, parent, message, identity)
// yields an identical SHA — re-runs are no-ops (force-push is idempotent, the prior run is reused).
const deterministicDate = "2000-01-01T00:00:00Z"

// CommitAll stages everything and makes a reproducible commit.
func CommitAll(dir, message string) error {
	if _, err := gitCapture([]string{"add", "-A"}, dir); err != nil {
		return err
	}
	env := append(os.Environ(), "GIT_AUTHOR_DATE="+deterministicDate, "GIT_COMMITTER_DATE="+deterministicDate)
	_, err := Capture("git", []string{"commit", "--allow-empty", "-m", message}, ExecOptions{Dir: dir, Env: env})
	return err
}

// ForcePush force-pushes the branch to origin.
func ForcePush(dir, branch string) error {
	_, err := gitCapture([]string{"push", "--force", "origin", branch}, dir)
	return err
}

// HeadSha returns the current HEAD commit SHA.
func HeadSha(dir string) (string, error) {
	out, err := gitCapture([]string{"rev-parse", "HEAD"}, dir)
	return strings.TrimSpace(out), err
}
