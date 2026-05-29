package adapters

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/stefanpenner-cs/reusable-workflows/internal/shadow/core"
)

// PatchWorkflowsInDir mirror-transforms the consumer's workflow files under
// <rootDir>/.github/workflows. Only files that actually call workflowsRepo are touched (leaving
// unrelated workflows alone so they aren't force-triggered on the shadow PR). Returns names changed.
func PatchWorkflowsInDir(rootDir, workflowsRepo, workflowsRef string) ([]string, error) {
	dir := filepath.Join(rootDir, ".github", "workflows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var changed []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		file := filepath.Join(dir, name)
		before, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		refs, err := core.ReferencesWorkflowsRepo(string(before), workflowsRepo)
		if err != nil {
			return nil, err
		}
		if !refs {
			continue
		}
		after, err := core.TransformWorkflowFile(string(before), workflowsRepo, workflowsRef)
		if err != nil {
			return nil, err
		}
		if after != string(before) {
			if err := os.WriteFile(file, []byte(after), 0o644); err != nil {
				return nil, err
			}
			changed = append(changed, name)
		}
	}
	return changed, nil
}
