package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireEnvReturnsValue(t *testing.T) {
	t.Setenv("REQUIRE_ENV_TEST", "value")
	v, err := RequireEnv("REQUIRE_ENV_TEST")
	require.NoError(t, err)
	assert.Equal(t, "value", v)
}

func TestRequireEnvErrorsWhenMissing(t *testing.T) {
	_, err := RequireEnv("REQUIRE_ENV_MISSING")
	assert.ErrorContains(t, err, "REQUIRE_ENV_MISSING")
}

func TestRequireEnvErrorsWhenEmpty(t *testing.T) {
	t.Setenv("REQUIRE_ENV_EMPTY", "")
	_, err := RequireEnv("REQUIRE_ENV_EMPTY")
	assert.Error(t, err)
}
