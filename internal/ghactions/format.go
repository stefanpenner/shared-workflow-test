// Package ghactions holds the GitHub Actions output helpers shared across the composite actions:
// section/group log formatting (ported from the old scripts/lib/log/format.mjs), the
// $GITHUB_OUTPUT sink, and named-flag validation. Every function is pure except the sink writer.
package ghactions

import "strings"

// Pair is an ordered key/value. Go maps don't preserve insertion order, so callers pass a slice
// to keep the deterministic column/line ordering the tests assert on.
type Pair struct {
	Key   string
	Value string
}

// Heading renders "▸ title".
func Heading(title string) string {
	return "▸ " + title
}

// KV aligns ordered pairs into "  key   value" rows, keys padded to the widest key.
func KV(pairs []Pair) string {
	width := 0
	for _, p := range pairs {
		if len(p.Key) > width {
			width = len(p.Key)
		}
	}
	rows := make([]string, len(pairs))
	for i, p := range pairs {
		rows[i] = "  " + p.Key + strings.Repeat(" ", width-len(p.Key)) + "  " + p.Value
	}
	return strings.Join(rows, "\n")
}

// Section is a heading line followed by aligned key/value rows.
func Section(title string, pairs []Pair) string {
	return Heading(title) + "\n" + KV(pairs)
}

// Group wraps a body in a collapsible GitHub Actions log group.
func Group(title, body string) string {
	return "::group::" + title + "\n" + body + "\n::endgroup::"
}
