package core

import "errors"

// Context is the resolved workflows PR number + head SHA to shadow-test.
type Context struct {
	PR  string
	SHA string
}

// ResolveContext decides which workflows PR + head SHA to shadow-test. On `pull_request` both come
// from the event; on `workflow_dispatch` only the PR is given, so the head SHA is looked up
// (lookupHeadSha is injected, keeping this pure/testable). Empty strings count as absent.
func ResolveContext(eventName, prNumber, headSha, inputPr string, lookupHeadSha func(pr string) (string, error)) (Context, error) {
	if eventName == "pull_request" {
		if prNumber == "" || headSha == "" {
			return Context{}, errors.New("pull_request needs --pr-number and --head-sha")
		}
		return Context{PR: prNumber, SHA: headSha}, nil
	}
	if inputPr == "" {
		return Context{}, errors.New("workflow_dispatch needs --input-pr")
	}
	sha, err := lookupHeadSha(inputPr)
	if err != nil {
		return Context{}, err
	}
	return Context{PR: inputPr, SHA: sha}, nil
}
