package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShadowBranchName(t *testing.T) {
	assert.Equal(t, "shadow/pr-7-stefanpenner-cs-reusable-workflows-consumer",
		ShadowBranchName(7, "stefanpenner-cs/reusable-workflows-consumer"))
}

func TestShadowBranchNameSlugifiesDotsAndSeparator(t *testing.T) {
	assert.Equal(t, "shadow/pr-1-org-lcc-live", ShadowBranchName(1, "org/lcc.live"))
}

func TestShadowBranchNameLowercases(t *testing.T) {
	assert.Equal(t, "shadow/pr-2-org-myrepo", ShadowBranchName(2, "Org/MyRepo"))
}

func TestShadowBranchNameKeepsOwner(t *testing.T) {
	assert.NotEqual(t, ShadowBranchName(1, "a/app"), ShadowBranchName(1, "b/app"))
}
