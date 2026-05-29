// Package adapters performs the shadow harness's real I/O: process exec, git, the GitHub API (via
// google/go-github), and consumer-workflow patching on disk. Excluded from the coverage gate (it's
// the I/O boundary); the pure logic it drives lives in internal/shadow/core.
package adapters

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ExecOptions configures Capture/Run.
type ExecOptions struct {
	Dir   string   // working directory ("" = inherit)
	Env   []string // full environment ("" = inherit)
	Input string   // written to stdin
}

// Capture runs a command and returns its stdout, erroring (with stderr) on a non-zero exit.
func Capture(name string, args []string, opts ExecOptions) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = opts.Dir
	cmd.Env = opts.Env
	if opts.Input != "" {
		cmd.Stdin = strings.NewReader(opts.Input)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("`%s %s` failed: %s", name, strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}

// Run streams a command's output (inherited stdio), erroring on a non-zero exit.
func Run(name string, args []string, opts ExecOptions) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = opts.Dir
	cmd.Env = opts.Env
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("`%s %s` exited non-zero: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
