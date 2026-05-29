package core

// TransformWorkflowFile is the full mirror transform for one consumer workflow: repoint the
// workflows ref, then guarantee a pull_request trigger so the shadow PR actually runs it.
func TransformWorkflowFile(yamlText, workflowsRepo, workflowsRef string) (string, error) {
	patched, err := PatchConsumerWorkflow(yamlText, workflowsRepo, workflowsRef)
	if err != nil {
		return "", err
	}
	return EnsurePullRequestTrigger(patched)
}
