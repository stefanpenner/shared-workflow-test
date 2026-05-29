package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const wfRepo = "stefanpenner-cs/reusable-workflows"
const wfSHA = "0123456789abcdef0123456789abcdef01234567"

func setupWorkflows(t *testing.T) (root, wfDir string) {
	t.Helper()
	root = t.TempDir()
	wfDir = filepath.Join(root, ".github", "workflows")
	require.NoError(t, os.MkdirAll(wfDir, 0o755))
	return root, wfDir
}

func writeWF(t *testing.T, wfDir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(wfDir, name), []byte(content), 0o644))
}

func readYAML(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, yaml.Unmarshal(b, &m))
	return m
}

func TestPatchWorkflowsRepointsAndTriggers(t *testing.T) {
	root, wfDir := setupWorkflows(t)
	writeWF(t, wfDir, "ci.yaml", "on: { push: { branches: [main] } }\njobs:\n  ci:\n    uses: stefanpenner-cs/reusable-workflows/.github/workflows/shared.yaml@main\n")

	changed, err := PatchWorkflowsInDir(root, wfRepo, wfSHA)
	require.NoError(t, err)
	assert.Equal(t, []string{"ci.yaml"}, changed)

	out := readYAML(t, filepath.Join(wfDir, "ci.yaml"))
	jobs := out["jobs"].(map[string]any)["ci"].(map[string]any)
	assert.Equal(t, wfRepo+"/.github/workflows/shared.yaml@"+wfSHA, jobs["uses"])
	assert.Equal(t, wfSHA, jobs["with"].(map[string]any)["ref"])
	_, hasPR := out["on"].(map[string]any)["pull_request"]
	assert.True(t, hasPR)
}

func TestPatchWorkflowsLeavesUnrelatedUntouched(t *testing.T) {
	root, wfDir := setupWorkflows(t)
	deploy := "on: { push: { tags: [\"v*\"] } }\njobs:\n  deploy:\n    runs-on: ubuntu-latest\n"
	writeWF(t, wfDir, "deploy.yaml", deploy)

	changed, err := PatchWorkflowsInDir(root, wfRepo, wfSHA)
	require.NoError(t, err)
	assert.Empty(t, changed)

	b, err := os.ReadFile(filepath.Join(wfDir, "deploy.yaml"))
	require.NoError(t, err)
	assert.Equal(t, deploy, string(b)) // byte-identical: never rewritten
}

func TestPatchWorkflowsIgnoresNonWorkflowFiles(t *testing.T) {
	root, wfDir := setupWorkflows(t)
	require.NoError(t, os.WriteFile(filepath.Join(wfDir, "notes.txt"), []byte("hello"), 0o644))
	writeWF(t, wfDir, "ci.yaml", "jobs:\n  ci:\n    uses: stefanpenner-cs/reusable-workflows/.github/workflows/shared.yaml@main\n")
	changed, err := PatchWorkflowsInDir(root, wfRepo, wfSHA)
	require.NoError(t, err)
	assert.Equal(t, []string{"ci.yaml"}, changed)
}

func TestPatchWorkflowsNoDir(t *testing.T) {
	changed, err := PatchWorkflowsInDir(t.TempDir(), wfRepo, wfSHA)
	require.NoError(t, err)
	assert.Empty(t, changed)
}
