// Package noinlinescripts enforces the "no inline scripts" rule: every action/workflow `run:` must
// be a single external invocation — a bazelisk/go/bash/sh call or a bare script path — never embedded
// shell logic, and `actions/github-script` is banned (it embeds inline JS). It parses the YAML and
// inspects each step's *folded* `run:` value, so flags split across continuation lines for
// readability are validated as one command (shell operators on a later line are still caught).
// Pure + tested; file discovery lives in tools/no-inline-scripts.
package noinlinescripts

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// AllowNames are step names exempt from the rule. Empty: shared.yaml bootstraps via
// stefanpenner/checkout-anywhere (a plain `uses:`), so there is no inline exception.
var AllowNames = map[string]bool{}

var (
	exprRe   = regexp.MustCompile(`\$\{\{[^}]*\}\}`)
	shellOps = regexp.MustCompile("&&|\\|\\||[;|`<>]|\\$\\(")
	evalRe   = regexp.MustCompile(`^(node|deno|bun)\s+(-e|--eval|-p|--print)\b`)
	// Accepted interpreters: bazelisk/bazel (the Go runtime model) + go (dev/CI tooling), plus bash/sh.
	interpRe = regexp.MustCompile(`^(go|bash|sh|bazelisk|bazel)\s+\S`)
	bareRe   = regexp.MustCompile(`^\S+\.(mjs|cjs|js|sh)$`)
	ghScript = regexp.MustCompile(`^actions/github-script@`)
)

// IsSingleInvocation reports whether value (the fully-folded run: command) is a single external
// invocation with no embedded logic.
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

// InlineErrors parses one YAML document and returns a Violation for every offending step `run:` (or
// banned `actions/github-script` use). allowNames (a set of exempt step names) is injectable for
// testing; pass AllowNames for the default policy.
func InlineErrors(yamlText string, allowNames map[string]bool) []Violation {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(yamlText), &doc); err != nil {
		return []Violation{{Line: 1, Message: "could not parse YAML: " + err.Error()}}
	}
	var out []Violation
	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		if n.Kind == yaml.MappingNode {
			name := ""
			for i := 0; i+1 < len(n.Content); i += 2 {
				if n.Content[i].Value == "name" && n.Content[i+1].Kind == yaml.ScalarNode {
					name = n.Content[i+1].Value
				}
			}
			for i := 0; i+1 < len(n.Content); i += 2 {
				key, val := n.Content[i], n.Content[i+1]
				switch {
				case key.Value == "uses" && val.Kind == yaml.ScalarNode && ghScript.MatchString(val.Value):
					out = append(out, Violation{val.Line, "actions/github-script embeds inline JS — write a tested external script instead"})
				case key.Value == "run" && val.Kind == yaml.ScalarNode && !allowNames[name]:
					switch {
					case val.Style == yaml.LiteralStyle || val.Style == yaml.FoldedStyle:
						out = append(out, Violation{val.Line, "block scalar run: — move logic into an external script"})
					case strings.TrimSpace(val.Value) == "":
						out = append(out, Violation{val.Line, "empty run: — nothing to invoke"})
					case !IsSingleInvocation(val.Value):
						out = append(out, Violation{val.Line, fmt.Sprintf("inline logic in run: %q — call an external script instead", val.Value)})
					}
				}
			}
		}
		for _, c := range n.Content {
			walk(c)
		}
	}
	walk(&doc)
	return out
}
