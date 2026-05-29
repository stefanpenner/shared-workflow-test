// Package test holds the pure logic for the Test action: render a section echoing the resolved
// suite + coverage state. (The action is a scaffold — see README; the real test run would slot in.)
package test

import (
	"strings"

	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
)

// Report renders the "▸ Test" section. Suite defaults to "unit"; coverage is enabled unless the
// value is (case-insensitively) "false".
func Report(suite, coverage string) string {
	s := strings.TrimSpace(suite)
	if s == "" {
		s = "unit"
	}
	state := "enabled"
	if strings.ToLower(strings.TrimSpace(coverage)) == "false" {
		state = "disabled"
	}
	return ghactions.Section("Test", []ghactions.Pair{
		{Key: "suite", Value: s},
		{Key: "coverage", Value: state},
	})
}
