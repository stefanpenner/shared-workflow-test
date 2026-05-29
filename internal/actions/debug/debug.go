// Package debug holds the pure diagnostics formatting for the Debug action: render the file tree
// and git status wrapped in collapsible GitHub Actions log groups. The exec function is injected
// so tests need no real filesystem or git; the cli main supplies the real implementation.
package debug

import (
	"errors"
	"strings"

	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
)

// ErrProbe marks an expected probe failure (command missing or non-zero exit). tryExec swallows it
// and substitutes fallback text; any other error propagates (it signals a bug, not a probe miss).
var ErrProbe = errors.New("probe failed")

// ExecFunc runs a command and returns its stdout, or an error (ErrProbe for expected failures).
type ExecFunc func(file string, args []string) (string, error)

func tryExec(exec ExecFunc, file string, args []string, fallback string) (string, error) {
	out, err := exec(file, args)
	switch {
	case err == nil:
		return out, nil
	case errors.Is(err, ErrProbe):
		return fallback, nil
	default:
		return "", err
	}
}

// TreeReport lists $HOME, $HOME/work/, and the project tree (minus .git/node_modules), wrapped in
// an "Environment" log group. env supplies HOME and GITHUB_WORKSPACE (global runner state).
func TreeReport(exec ExecFunc, env map[string]string) (string, error) {
	home := env["HOME"]
	workspace := env["GITHUB_WORKSPACE"]
	if workspace == "" {
		workspace = "."
	}
	homeLs, err := tryExec(exec, "ls", []string{"-la", home}, "(unavailable)")
	if err != nil {
		return "", err
	}
	workLs, err := tryExec(exec, "ls", []string{"-la", home + "/work/"}, "(not found)")
	if err != nil {
		return "", err
	}
	tree, err := tryExec(exec, "find", []string{workspace, "-not", "-path", "*/.git/*", "-not", "-path", "*/node_modules/*"}, "(unavailable)")
	if err != nil {
		return "", err
	}
	body := strings.Join([]string{
		"$HOME:", homeLs, "",
		"$HOME/work/:", workLs, "",
		"project tree:", tree,
	}, "\n")
	return ghactions.Group("Environment", body), nil
}

// GitReport renders git status + diffs in a "Git status" group, or a "no repository" note when the
// working directory isn't a git repo.
func GitReport(exec ExecFunc) (string, error) {
	if _, err := exec("git", []string{"rev-parse", "--git-dir"}); err != nil {
		if errors.Is(err, ErrProbe) {
			return ghactions.Group("Git status", "No git repository in working directory"), nil
		}
		return "", err
	}
	status, err := exec("git", []string{"status"})
	if err != nil {
		return "", err
	}
	diff, err := exec("git", []string{"diff"})
	if err != nil {
		return "", err
	}
	cached, err := exec("git", []string{"diff", "--cached"})
	if err != nil {
		return "", err
	}
	body := strings.Join([]string{
		status, "",
		"unstaged changes:", diff, "",
		"staged changes:", cached,
	}, "\n")
	return ghactions.Group("Git status", body), nil
}
