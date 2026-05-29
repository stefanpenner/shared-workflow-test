// Command guard is the no-inline-scripts check: discover every action + workflow YAML, run the
// guard, and exit non-zero on any violation. Invoked from CI (test.yaml). Discovery (the old
// check-no-inline-scripts.cli.mjs) lives here; the rules live in internal/guard.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stefanpenner-cs/reusable-workflows/internal/guard"
)

// listDir returns directory entry names, tolerating only a missing directory.
func listDir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		panic(err)
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names
}

// discover finds actions/*/action.yaml and .github/workflows/*.{yaml,yml}.
func discover() []string {
	var files []string
	for _, name := range listDir("actions") {
		p := filepath.Join("actions", name, "action.yaml")
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			files = append(files, p)
		}
	}
	for _, name := range listDir(".github/workflows") {
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, filepath.Join(".github/workflows", name))
		}
	}
	return files
}

func main() {
	// `bazel run` executes in the runfiles tree; BUILD_WORKSPACE_DIRECTORY points back at the repo
	// root so discovery sees the real actions/ and .github/workflows/.
	if ws := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); ws != "" {
		if err := os.Chdir(ws); err != nil {
			fmt.Fprintf(os.Stderr, "could not enter workspace %s: %v\n", ws, err)
			os.Exit(1)
		}
	}

	files := discover()
	violations := 0
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not read %s for inline-script check: %v\n", f, err)
			os.Exit(1)
		}
		for _, v := range guard.InlineErrors(string(content), guard.AllowNames) {
			fmt.Fprintf(os.Stderr, "✗ %s:%d  %s\n", f, v.Line, v.Message)
			violations++
		}
	}
	if violations > 0 {
		fmt.Fprintf(os.Stderr, "\n✗ no-inline-scripts: %d violation(s) across %d file(s)\n", violations, len(files))
		os.Exit(1)
	}
	fmt.Printf("✓ no-inline-scripts: %d file(s) clean\n", len(files))
}
