package core

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
	edgeDash = regexp.MustCompile(`^-+|-+$`)
)

// ShadowBranchName builds a deterministic, collision-safe branch name for a shadow PR:
// `shadow/pr-<n>-<slug>`. Deterministic so the workflows find the runner PR without discovery and
// re-runs reuse the branch (force-push → synchronize). The owner is kept in the slug so same-named
// repos under different owners don't collide.
func ShadowBranchName(prNumber int, consumerRepo string) string {
	slug := strings.ToLower(consumerRepo)
	slug = nonAlnum.ReplaceAllString(slug, "-")
	slug = edgeDash.ReplaceAllString(slug, "")
	return fmt.Sprintf("shadow/pr-%d-%s", prNumber, slug)
}
