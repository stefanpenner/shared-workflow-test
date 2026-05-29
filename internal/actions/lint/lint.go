// Package lint holds the pure logic for the Lint action: render a section echoing the resolved
// paths/config. (The action is a scaffold — see README; real linting would slot in here.)
package lint

import (
	"strings"

	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
)

// Report renders the "▸ Lint" section, defaulting paths to "." and config to ".eslintrc".
func Report(paths, config string) string {
	p := strings.TrimSpace(paths)
	if p == "" {
		p = "."
	}
	c := strings.TrimSpace(config)
	if c == "" {
		c = ".eslintrc"
	}
	return ghactions.Section("Lint", []ghactions.Pair{
		{Key: "paths", Value: p},
		{Key: "config", Value: c},
	})
}
