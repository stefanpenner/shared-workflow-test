// Command wsrun runs a hermetic tool against the source tree. `bazel run` starts a binary in the
// runfiles tree, so lint tools like yamllint/golangci-lint would see copied inputs, not the repo.
// wsrun cd's to BUILD_WORKSPACE_DIRECTORY (the dir `bazel run` was invoked from) and execs the tool
// — located via its runfiles rlocationpath, passed as the first arg — with the remaining args. It
// is the non-Go analogue of the os.Chdir(BUILD_WORKSPACE_DIRECTORY) dance our Go tools already do.
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "wsrun: usage: wsrun <tool-rlocationpath> [args...]")
		os.Exit(2)
	}
	tool, err := runfiles.Rlocation(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "wsrun: locating tool:", err)
		os.Exit(1)
	}
	ws := os.Getenv("BUILD_WORKSPACE_DIRECTORY")
	if ws == "" {
		fmt.Fprintln(os.Stderr, "wsrun: BUILD_WORKSPACE_DIRECTORY unset — run via `bazel run`")
		os.Exit(1)
	}

	cmd := exec.Command(tool, os.Args[2:]...)
	cmd.Dir = ws
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "wsrun:", err)
		os.Exit(1)
	}
}
