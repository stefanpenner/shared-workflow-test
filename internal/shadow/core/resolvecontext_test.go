package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noLookup(string) (string, error) { return "", errors.New("should not look up") }

func TestResolveContextPullRequest(t *testing.T) {
	ctx, err := ResolveContext("pull_request", "7", "abc", "", noLookup)
	require.NoError(t, err)
	assert.Equal(t, Context{PR: "7", SHA: "abc"}, ctx)
}

func TestResolveContextWorkflowDispatchLooksUpSHA(t *testing.T) {
	ctx, err := ResolveContext("workflow_dispatch", "", "", "9", func(pr string) (string, error) {
		return "sha-of-" + pr, nil
	})
	require.NoError(t, err)
	assert.Equal(t, Context{PR: "9", SHA: "sha-of-9"}, ctx)
}

func TestResolveContextPullRequestMissingSHA(t *testing.T) {
	_, err := ResolveContext("pull_request", "7", "", "", noLookup)
	assert.ErrorContains(t, err, "head-sha")
}

func TestResolveContextDispatchMissingPR(t *testing.T) {
	_, err := ResolveContext("workflow_dispatch", "", "", "", noLookup)
	assert.ErrorContains(t, err, "input-pr")
}
