// Package setup holds the pure logic for the Setup action: validate the node version and render
// the log section. No I/O, no env reads — the cobra main wires those.
package setup

import (
	"errors"
	"strings"

	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
)

// ResolveNodeVersion trims the requested version and errors if it's empty.
func ResolveNodeVersion(input string) (string, error) {
	v := strings.TrimSpace(input)
	if v == "" {
		return "", errors.New("node-version is required")
	}
	return v, nil
}

// Report renders the "▸ Setup" section, falling back to "(unknown project)" for an empty name.
func Report(projectName, nodeVersion string) string {
	name := strings.TrimSpace(projectName)
	if name == "" {
		name = "(unknown project)"
	}
	return ghactions.Section("Setup", []ghactions.Pair{
		{Key: "project", Value: name},
		{Key: "node version", Value: nodeVersion},
	})
}
