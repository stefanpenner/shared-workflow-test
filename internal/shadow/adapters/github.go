package adapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"

	"github.com/stefanpenner-cs/reusable-workflows/internal/shadow/core"
)

const receiverWorkflow = "receiver.yaml"

func ptr[T any](v T) *T { return &v }

func newClient(token string) *github.Client { return github.NewClient(nil).WithAuthToken(token) }

func split(repo string) (owner, name string) {
	owner, name, _ = strings.Cut(repo, "/")
	return owner, name
}

// DispatchReceiver triggers the runner receiver via workflow_dispatch and returns the created run
// id (via the `return_run_details:true` capability — no run-discovery polling needed).
func DispatchReceiver(runnerRepo string, ctx core.ShadowContext, token string) (int, error) {
	return dispatchReceiver(newClient(token), runnerRepo, ctx)
}

func dispatchReceiver(cl *github.Client, runnerRepo string, ctx core.ShadowContext) (int, error) {
	body := map[string]any{
		"ref":                "main",
		"inputs":             core.BuildDispatchInputs(ctx),
		"return_run_details": true,
	}
	u := fmt.Sprintf("repos/%s/actions/workflows/%s/dispatches", runnerRepo, receiverWorkflow)
	req, err := cl.NewRequest("POST", u, body)
	if err != nil {
		return 0, err
	}
	var result map[string]any
	if _, err := cl.Do(context.Background(), req, &result); err != nil {
		return 0, err
	}
	return core.ExtractRunID(result)
}

// PullRequestHeadSHA returns the head commit SHA of a PR (used to resolve the SHA on a
// workflow_dispatch, where only the PR number is given).
func PullRequestHeadSHA(repo, prNumber, token string) (string, error) {
	owner, name := split(repo)
	n, err := strconv.Atoi(prNumber)
	if err != nil {
		return "", fmt.Errorf("invalid PR number %q: %w", prNumber, err)
	}
	pr, _, err := newClient(token).PullRequests.Get(context.Background(), owner, name, n)
	if err != nil {
		return "", err
	}
	return pr.GetHead().GetSHA(), nil
}

// WatchRun polls the receiver run to completion (quietly); errors on a non-success conclusion.
func WatchRun(runnerRepo string, runID int, token string) error {
	return awaitRun(newClient(token), runnerRepo, int64(runID), "receiver run", 180, 5*time.Second)
}

func awaitRun(cl *github.Client, repo string, runID int64, label string, attempts int, interval time.Duration) error {
	owner, name := split(repo)
	for i := 0; i < attempts; i++ {
		run, _, err := cl.Actions.GetWorkflowRunByID(context.Background(), owner, name, runID)
		if err != nil {
			return err
		}
		switch core.ClassifyRunState(run.GetStatus(), run.GetConclusion()) {
		case core.RunSuccess:
			return nil
		case core.RunFailure:
			concl := run.GetConclusion()
			if concl == "" {
				concl = "failed"
			}
			return fmt.Errorf("%s #%d %s", label, runID, concl)
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("timed out waiting for %s #%d after %s", label, runID, time.Duration(attempts)*interval)
}

func findOpenPR(cl *github.Client, repo, branch string) (*github.PullRequest, error) {
	owner, name := split(repo)
	prs, _, err := cl.PullRequests.List(context.Background(), owner, name, &github.PullRequestListOptions{
		State: "open",
		Head:  owner + ":" + branch,
	})
	if err != nil {
		return nil, err
	}
	if len(prs) == 0 {
		return nil, nil
	}
	return prs[0], nil
}

// FindPrURL returns the URL of the open PR for a head branch, or "" if none exists.
func FindPrURL(repo, branch, token string) (string, error) {
	return findPrURL(newClient(token), repo, branch)
}

func findPrURL(cl *github.Client, repo, branch string) (string, error) {
	pr, err := findOpenPR(cl, repo, branch)
	if err != nil || pr == nil {
		return "", err
	}
	return pr.GetHTMLURL(), nil
}

// EnsurePR opens the shadow PR if one isn't already open for the branch; returns its URL.
func EnsurePR(repo, branch, base, title, body, token string) (string, error) {
	return ensurePR(newClient(token), repo, branch, base, title, body)
}

func ensurePR(cl *github.Client, repo, branch, base, title, body string) (string, error) {
	if existing, err := findPrURL(cl, repo, branch); err != nil {
		return "", err
	} else if existing != "" {
		return existing, nil
	}
	owner, name := split(repo)
	pr, _, err := cl.PullRequests.Create(context.Background(), owner, name, &github.NewPullRequest{
		Title: ptr(title),
		Head:  ptr(branch),
		Base:  ptr(base),
		Body:  ptr(body),
	})
	if err != nil {
		return "", err
	}
	return pr.GetHTMLURL(), nil
}

// WatchCommitRun waits for the workflow run of an exact commit SHA to finish; errors on failure.
// Keying on the SHA (not the branch) is deterministic — immune to stale checks and the "no checks
// yet" race. The only inherent wait is GitHub creating the run for the just-pushed SHA.
func WatchCommitRun(repo, sha, token string) error {
	return watchCommitRun(newClient(token), repo, sha, 40, 5*time.Second, 180, 5*time.Second)
}

func watchCommitRun(cl *github.Client, repo, sha string, findAttempts int, findInterval time.Duration, runAttempts int, runInterval time.Duration) error {
	owner, name := split(repo)
	var runID int64
	for i := 0; i < findAttempts; i++ {
		runs, _, err := cl.Actions.ListRepositoryWorkflowRuns(context.Background(), owner, name, &github.ListWorkflowRunsOptions{HeadSHA: sha})
		if err == nil && runs != nil && len(runs.WorkflowRuns) > 0 {
			runID = runs.WorkflowRuns[0].GetID()
		}
		if runID != 0 {
			break
		}
		if i == findAttempts-1 {
			return fmt.Errorf("no workflow run appeared for %s@%s after %s", repo, sha, time.Duration(findAttempts)*findInterval)
		}
		time.Sleep(findInterval)
	}
	return awaitRun(cl, repo, runID, "consumer CI", runAttempts, runInterval)
}

// ProbeDispatchStatus POSTs the receiver's workflow_dispatch endpoint with the given Authorization
// header value ("" sends no Authorization header) and returns the HTTP status code. It is the I/O
// half of the dispatch-auth security check: it deliberately sends only a credential we expect to be
// rejected and a body that GitHub rejects before ever reading, so it can NEVER trigger a real run.
// It never reads the ambient GITHUB_TOKEN, so running it inside Actions is safe.
func ProbeDispatchStatus(runnerRepo, authHeader string) (int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows/%s/dispatches", runnerRepo, receiverWorkflow)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, strings.NewReader(`{}`))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode, nil
}

// ClosePRAndDeleteBranch closes the shadow PR (if any) and deletes its branch. Best-effort: a PR or
// branch that's already gone is fine.
func ClosePRAndDeleteBranch(repo, branch, token string) error {
	return closePRAndDeleteBranch(newClient(token), repo, branch)
}

func closePRAndDeleteBranch(cl *github.Client, repo, branch string) error {
	owner, name := split(repo)
	if pr, err := findOpenPR(cl, repo, branch); err == nil && pr != nil {
		_, _, _ = cl.PullRequests.Edit(context.Background(), owner, name, pr.GetNumber(), &github.PullRequest{State: ptr("closed")})
	}
	_, _ = cl.Git.DeleteRef(context.Background(), owner, name, "heads/"+branch)
	return nil
}
