// Package core holds the pure shadow-testing logic (branch naming, consumer parsing, dispatch
// classification, summary rendering, and the comment-preserving workflow transforms). No process
// I/O except RequireEnv; the git/GitHub/filesystem work lives in internal/shadow/adapters.
package core

import (
	"fmt"
	"os"
)

// RequireEnv reads a required environment variable, erroring (naming it) when absent or empty.
func RequireEnv(name string) (string, error) {
	if v := os.Getenv(name); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("missing required environment variable: %s", name)
}
