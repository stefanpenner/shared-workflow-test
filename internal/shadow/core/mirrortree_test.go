package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMirrorTreeCopiesRecursively(t *testing.T) {
	src, dest := t.TempDir(), t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(src, ".github", "workflows"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "README.md"), []byte("hi"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(src, ".github", "workflows", "ci.yaml"), []byte("on: push"), 0o644))

	require.NoError(t, MirrorTree(src, dest))

	b, err := os.ReadFile(filepath.Join(dest, "README.md"))
	require.NoError(t, err)
	assert.Equal(t, "hi", string(b))
	assert.FileExists(t, filepath.Join(dest, ".github", "workflows", "ci.yaml"))
}

func TestMirrorTreeNeverCopiesGit(t *testing.T) {
	src, dest := t.TempDir(), t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(src, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, ".git", "HEAD"), []byte("ref: refs/heads/main"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "app.js"), []byte("x"), 0o644))

	require.NoError(t, MirrorTree(src, dest))

	assert.FileExists(t, filepath.Join(dest, "app.js"))
	assert.NoDirExists(t, filepath.Join(dest, ".git"))
}
