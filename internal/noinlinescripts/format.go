package noinlinescripts

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var runLineRe = regexp.MustCompile(`^(\s*)(- )?run: (\S.*)$`)

// Format rewrites any single-line `run:` that is a single invocation with 2+ long flags into the
// multi-line form — program (and the `--` separator) on the `run:` line, each `--flag` on its own
// continuation line indented +2. Returns the formatted text and whether it changed. Idempotent.
// Only genuine step `run:` scalars (located via the YAML parse) are touched, so a line that merely
// looks like `run:` inside other content is never rewritten.
func Format(yamlText string) (string, bool) {
	targets := splitTargets(yamlText)
	if len(targets) == 0 {
		return yamlText, false
	}
	lines := strings.Split(yamlText, "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for i, line := range lines {
		if targets[i+1] {
			if split, ok := splitRunLine(line); ok {
				out = append(out, split...)
				changed = true
				continue
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n"), changed
}

// splitTargets returns the 1-based line numbers of step `run:` scalars that are a single invocation
// with 2+ long flags crammed on the line (the lines Format should split).
func splitTargets(yamlText string) map[int]bool {
	var doc yaml.Node
	if yaml.Unmarshal([]byte(yamlText), &doc) != nil {
		return nil
	}
	lines := strings.Split(yamlText, "\n")
	targets := map[int]bool{}
	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		if n.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(n.Content); i += 2 {
				k, v := n.Content[i], n.Content[i+1]
				if k.Value == "run" && v.Kind == yaml.ScalarNode &&
					v.Style != yaml.LiteralStyle && v.Style != yaml.FoldedStyle &&
					IsSingleInvocation(v.Value) &&
					v.Line-1 < len(lines) && len(flagRe.FindAllString(lines[v.Line-1], -1)) >= 2 {
					targets[v.Line] = true
				}
			}
		}
		for _, c := range n.Content {
			walk(c)
		}
	}
	walk(&doc)
	return targets
}

// splitRunLine rewrites one crammed `run:` line into [run: <program> --, <flag>, <flag>, …].
func splitRunLine(line string) ([]string, bool) {
	m := runLineRe.FindStringSubmatch(line)
	if m == nil {
		return nil, false
	}
	indent, dash, raw := m[1], m[2], m[3]
	tokens := tokenize(stripQuotes(strings.TrimSpace(raw)))

	// Split at the first real flag (a token starting with "-" that isn't the bare "--" separator).
	idx := -1
	for i, t := range tokens {
		if strings.HasPrefix(t, "-") && t != "--" {
			idx = i
			break
		}
	}
	if idx < 1 {
		return nil, false // no program prefix, or no flags — leave it
	}

	contIndent := strings.Repeat(" ", len(indent)+len(dash)+2)
	result := []string{indent + dash + "run: " + strings.Join(tokens[:idx], " ")}
	for _, flag := range tokens[idx:] {
		result = append(result, contIndent+flag)
	}
	return result, true
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// tokenize splits a command on whitespace, keeping quoted spans and `${{ … }}` expressions atomic
// (GHA expressions contain unquoted spaces that must not be split on).
func tokenize(s string) []string {
	var tokens []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	flush := func() {
		if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}
	for i := 0; i < len(s); i++ {
		if !inSingle && !inDouble && strings.HasPrefix(s[i:], "${{") {
			if end := strings.Index(s[i:], "}}"); end >= 0 {
				cur.WriteString(s[i : i+end+2])
				i += end + 1
				continue
			}
		}
		c := s[i]
		switch {
		case c == '\'' && !inDouble:
			inSingle = !inSingle
			cur.WriteByte(c)
		case c == '"' && !inSingle:
			inDouble = !inDouble
			cur.WriteByte(c)
		case (c == ' ' || c == '\t') && !inSingle && !inDouble:
			flush()
		default:
			cur.WriteByte(c)
		}
	}
	flush()
	return tokens
}
