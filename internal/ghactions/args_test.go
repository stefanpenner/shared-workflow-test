package ghactions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequireFlagsAcceptsNonEmpty(t *testing.T) {
	assert.NoError(t, RequireFlags([]Pair{{"paths", "src"}, {"config", ".x"}}))
}

func TestRequireFlagsNamesTheEmptyFlag(t *testing.T) {
	err := RequireFlags([]Pair{{"project-name", ""}})
	assert.EqualError(t, err, "missing required --project-name")
}
