package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConsumers(t *testing.T) {
	got, err := ParseConsumers(`[{"repo":"stefanpenner-cs/reusable-workflows-consumer","ref":"main"}]`)
	require.NoError(t, err)
	assert.Equal(t, []Consumer{{Repo: "stefanpenner-cs/reusable-workflows-consumer", Ref: "main"}}, got)
}

func TestParseConsumersDefaultsRef(t *testing.T) {
	got, err := ParseConsumers(`[{"repo":"o/r"}]`)
	require.NoError(t, err)
	assert.Equal(t, []Consumer{{Repo: "o/r", Ref: "main"}}, got)
}

func TestParseConsumersAllowsEmpty(t *testing.T) {
	got, err := ParseConsumers(`[]`)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestParseConsumersRejectsBadInput(t *testing.T) {
	_, err := ParseConsumers("not json")
	assert.Error(t, err)
	_, err = ParseConsumers(`[{"ref":"main"}]`) // repo missing
	assert.Error(t, err)
	_, err = ParseConsumers(`[{"repo":"nope"}]`) // not owner/name
	assert.Error(t, err)
	_, err = ParseConsumers(`{"repo":"o/r"}`) // not an array
	assert.Error(t, err)
}
