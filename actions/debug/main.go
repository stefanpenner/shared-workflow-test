// Command debug is the Debug action entrypoint: print the file tree + git status. It supplies the
// real exec (stdout captured, stderr discarded; failures become ErrProbe so probes stay quiet).
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/stefanpenner-cs/reusable-workflows/internal/actions/debug"
)

// realExec runs a command, returning stdout. A missing command or non-zero exit becomes ErrProbe
// (an expected probe failure), so the formatters substitute fallback text instead of erroring.
func realExec(file string, args []string) (string, error) {
	out, err := exec.Command(file, args...).Output() // stderr discarded
	if err != nil {
		return "", fmt.Errorf("%w: %s: %v", debug.ErrProbe, file, err)
	}
	return string(out), nil
}

func main() {
	cmd := &cobra.Command{
		Use:           "debug",
		Short:         "Print file tree + git status (diagnostics)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			env := map[string]string{"HOME": os.Getenv("HOME"), "GITHUB_WORKSPACE": os.Getenv("GITHUB_WORKSPACE")}
			tree, err := debug.TreeReport(realExec, env)
			if err != nil {
				return err
			}
			fmt.Println(tree)
			git, err := debug.GitReport(realExec)
			if err != nil {
				return err
			}
			fmt.Println(git)
			return nil
		},
	}
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
