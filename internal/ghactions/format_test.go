package ghactions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeading(t *testing.T) {
	assert.Equal(t, "▸ Setup", Heading("Setup"))
}

func TestKVAlignsToWidestKey(t *testing.T) {
	got := KV([]Pair{{"project", "demo"}, {"node version", "20"}})
	assert.Equal(t, "  project       demo\n  node version  20", got)
}

func TestKVEmpty(t *testing.T) {
	assert.Equal(t, "", KV(nil))
}

func TestSection(t *testing.T) {
	got := Section("Setup", []Pair{{"project", "demo"}, {"node version", "20"}})
	assert.Equal(t, "▸ Setup\n  project       demo\n  node version  20", got)
}

func TestGroup(t *testing.T) {
	assert.Equal(t, "::group::Environment\nbody\n::endgroup::", Group("Environment", "body"))
}
