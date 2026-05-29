package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDispatchRejected(t *testing.T) {
	// Refusals: missing/invalid credential, lacks actions:write, can't see the repo.
	for _, s := range []int{401, 403, 404} {
		assert.True(t, DispatchRejected(s), "status %d should count as refused", s)
	}
	// Not refusals: 204 is the dispatch success code (a real trigger), 200/500 are not refusals.
	for _, s := range []int{200, 204, 422, 500} {
		assert.False(t, DispatchRejected(s), "status %d should NOT count as refused", s)
	}
}

func TestLeakedProbes(t *testing.T) {
	all := []DispatchProbe{
		{Name: "unauthenticated", Status: 401},
		{Name: "invalid-token", Status: 401},
	}
	assert.Empty(t, LeakedProbes(all), "all-refused should leak nothing")

	mixed := []DispatchProbe{
		{Name: "unauthenticated", Status: 401},
		{Name: "somehow-accepted", Status: 204},
	}
	leaked := LeakedProbes(mixed)
	assert.Equal(t, []DispatchProbe{{Name: "somehow-accepted", Status: 204}}, leaked)
}
