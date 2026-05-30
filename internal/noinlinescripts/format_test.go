package noinlinescripts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// runValue parses a `steps:` doc and returns the first step's folded run: value.
func runValue(t *testing.T, s string) string {
	t.Helper()
	var m struct {
		Steps []map[string]string
	}
	require.NoError(t, yaml.Unmarshal([]byte(s), &m))
	require.NotEmpty(t, m.Steps)
	return m.Steps[0]["run"]
}

func TestFormatSplitsActionRunAndPreservesMeaning(t *testing.T) {
	in := "steps:\n  - run: \"bazelisk run //actions/setup -- --project-name='${{ inputs.project-name }}' --node-version='${{ inputs.node-version }}'\"\n"
	out, changed := Format(in)
	assert.True(t, changed)
	assert.Equal(t, runValue(t, in), runValue(t, out), "formatted output must fold to the same command")
	assert.Contains(t, out, "  - run: bazelisk run //actions/setup --\n")
	assert.Contains(t, out, "\n      --project-name='${{ inputs.project-name }}'\n")
	assert.Contains(t, out, "\n      --node-version='${{ inputs.node-version }}'\n")
}

func TestFormatKeepsUnquotedExpressionsWithSpacesIntact(t *testing.T) {
	in := "steps:\n  - run: bazelisk run //shadow/cmd/x -- --workflows-repo=${{ github.repository }} --input-pr=${{ inputs.pr }}\n"
	out, changed := Format(in)
	assert.True(t, changed)
	assert.Equal(t, runValue(t, in), runValue(t, out))
	assert.Contains(t, out, "\n      --workflows-repo=${{ github.repository }}\n")
	assert.Contains(t, out, "\n      --input-pr=${{ inputs.pr }}\n")
}

func TestFormatLeavesSingleFlagAlone(t *testing.T) {
	_, changed := Format("steps:\n  - run: bazelisk test //... --config=ci\n")
	assert.False(t, changed)
}

func TestFormatIsIdempotent(t *testing.T) {
	in := "steps:\n  - run: bazelisk run //actions/lint -- --paths='${{ inputs.paths }}' --config='${{ inputs.config }}'\n"
	once, changed := Format(in)
	require.True(t, changed)
	twice, changedAgain := Format(once)
	assert.False(t, changedAgain)
	assert.Equal(t, once, twice)
}

func TestFormatDoesNotTouchRunLikeBlockScalarContent(t *testing.T) {
	// `run: …` inside a block scalar is not a step run: — must be left byte-for-byte alone.
	in := "description: |\n  run: foo --a --b\njobs: {}\n"
	out, changed := Format(in)
	assert.False(t, changed)
	assert.Equal(t, in, out)
}
