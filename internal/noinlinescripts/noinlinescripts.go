// Package noinlinescripts enforces the "no inline scripts" rule: every action/workflow `run:` must be a
// single external invocation — a bazelisk/node/bash/sh call or a bare script path — never embedded
// shell logic, and `actions/github-script` is banned (it embeds inline JS). Ported from the old
// scripts/lib/guard/check-no-inline-scripts.mjs; pure + tested. File discovery lives in tools/guard.
package noinlinescripts

import (
	"fmt"
	"regexp"
	"strings"
)

// AllowNames are step names exempt from the rule. Empty: shared.yaml bootstraps via
// stefanpenner/checkout-anywhere (a plain `uses:`), so there is no inline exception.
var AllowNames = map[string]bool{}

var (
	exprRe   = regexp.MustCompile(`\$\{\{[^}]*\}\}`)
	shellOps = regexp.MustCompile("&&|\\|\\||[;|`<>]|\\$\\(")
	evalRe   = regexp.MustCompile(`^(node|deno|bun)\s+(-e|--eval|-p|--print)\b`)
	// Accepted interpreters: bazelisk/bazel (the Go runtime model) + go (CI tooling), plus bash/sh.
	interpRe = regexp.MustCompile(`^(go|bash|sh|bazelisk|bazel)\s+\S`)
	bareRe   = regexp.MustCompile(`^\S+\.(mjs|cjs|js|sh)$`)
	nameRe   = regexp.MustCompile(`^\s*-?\s*name:\s*(.+?)\s*$`)
	usesRe   = regexp.MustCompile(`^\s*-?\s*uses:\s*(.+?)\s*$`)
	runRe    = regexp.MustCompile(`^\s*-?\s*run:\s*(.*)$`)
	blockRe  = regexp.MustCompile(`^[|>][+-]?\d*$`)
	ghScript = regexp.MustCompile(`^actions/github-script@`)
)

func unquote(text string) string {
	t := strings.TrimSpace(text)
	if len(t) >= 2 {
		if (t[0] == '"' && t[len(t)-1] == '"') || (t[0] == '\'' && t[len(t)-1] == '\'') {
			return t[1 : len(t)-1]
		}
	}
	return t
}

// IsSingleInvocation reports whether value is a single external invocation with no embedded logic.
func IsSingleInvocation(value string) bool {
	v := strings.TrimSpace(exprRe.ReplaceAllString(value, "X")) // drop ${{ … }} before inspecting
	switch {
	case v == "":
		return false
	case shellOps.MatchString(v):
		return false
	case evalRe.MatchString(v): // inline eval defeats the rule even without shell operators
		return false
	case interpRe.MatchString(v):
		return true
	default:
		return bareRe.MatchString(v)
	}
}

// Violation is a single guard failure: a 1-based line number and a message.
type Violation struct {
	Line    int
	Message string
}

// InlineErrors scans one YAML document and returns a Violation per offending line. allowNames (a
// set of exempt step names) is injectable for testing; pass AllowNames for the default policy.
func InlineErrors(yamlText string, allowNames map[string]bool) []Violation {
	lines := strings.Split(yamlText, "\n")
	var out []Violation
	lastName := ""
	for i, line := range lines {
		if m := nameRe.FindStringSubmatch(line); m != nil {
			lastName = unquote(m[1])
		}
		if m := usesRe.FindStringSubmatch(line); m != nil && ghScript.MatchString(unquote(m[1])) {
			out = append(out, Violation{i + 1, "actions/github-script embeds inline JS — write a tested external script instead"})
			continue
		}
		m := runRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		allowed := allowNames[lastName]
		lastName = "" // a name applies to one step; don't let a later unnamed run inherit it
		if allowed {
			continue
		}
		raw := strings.TrimSpace(m[1])
		if raw == "" || blockRe.MatchString(raw) {
			out = append(out, Violation{i + 1, "block scalar run: — move logic into an external script"})
			continue
		}
		if value := unquote(raw); !IsSingleInvocation(value) {
			out = append(out, Violation{i + 1, fmt.Sprintf("inline logic in run: %q — call an external script instead", value)})
		}
	}
	return out
}
