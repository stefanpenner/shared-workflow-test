package ghactions

import "fmt"

// RequireFlags errors, naming the first flag whose value is empty. cobra parses the flags and
// enforces presence; this adds the "reject empty value" check the old requireArgs guaranteed, so
// `--flag=` fails loudly with the flag name (CLAUDE.md rule: CLI args, validated, never env).
func RequireFlags(flags []Pair) error {
	for _, f := range flags {
		if f.Value == "" {
			return fmt.Errorf("missing required --%s", f.Key)
		}
	}
	return nil
}
