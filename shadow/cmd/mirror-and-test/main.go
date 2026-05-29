// Command mirror-and-test (receiver, on workflow_dispatch): mirror the consumer's code onto a
// shadow branch with its workflows repointed at the workflows PR SHA, open/refresh a real PR, and
// block on that PR's checks — so this run's exit status IS the shadow-test result.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
	"github.com/stefanpenner-cs/reusable-workflows/internal/shadow/adapters"
	"github.com/stefanpenner-cs/reusable-workflows/internal/shadow/core"
)

func run(workflowsRepo, workflowsRef, consumerRepo, consumerRef, workflowsPR, branch, runnerRepo string) error {
	token, err := core.RequireEnv("SHADOW_PAT")
	if err != nil {
		return err
	}

	work, err := os.MkdirTemp("", "shadow-")
	if err != nil {
		return err
	}
	mirrorDir := filepath.Join(work, "consumer")
	shadowDir := filepath.Join(work, "runner")

	if err := adapters.CloneShallow(consumerRepo, consumerRef, mirrorDir, token); err != nil {
		return err
	}
	if err := adapters.CloneShallow(runnerRepo, "main", shadowDir, token); err != nil {
		return err
	}
	if err := adapters.ConfigureBotIdentity(shadowDir); err != nil {
		return err
	}
	if err := adapters.ResetBranchToEmptyTree(shadowDir, branch); err != nil {
		return err
	}
	if err := core.MirrorTree(mirrorDir, shadowDir); err != nil {
		return err
	}

	patched, err := adapters.PatchWorkflowsInDir(shadowDir, workflowsRepo, workflowsRef)
	if err != nil {
		return err
	}
	if len(patched) == 0 {
		fmt.Println("patched workflows: (none — consumer has no workflows?)")
	} else {
		fmt.Printf("patched workflows: %s\n", strings.Join(patched, ", "))
	}

	msg := fmt.Sprintf("shadow: %s@%s vs %s@%s (workflows PR #%s)", consumerRepo, consumerRef, workflowsRepo, workflowsRef, workflowsPR)
	if err := adapters.CommitAll(shadowDir, msg); err != nil {
		return err
	}
	if err := adapters.ForcePush(shadowDir, branch); err != nil {
		return err
	}
	sha, err := adapters.HeadSha(shadowDir)
	if err != nil {
		return err
	}

	body := strings.Join([]string{
		"Automated shadow test — **do not merge**.",
		"",
		fmt.Sprintf("- Consumer: `%s@%s`", consumerRepo, consumerRef),
		fmt.Sprintf("- Workflows draft: `%s@%s`", workflowsRepo, workflowsRef),
		fmt.Sprintf("- Workflows PR: %s#%s", workflowsRepo, workflowsPR),
		"",
		"This PR exists only to run the consumer's CI under a real `pull_request` event.",
	}, "\n")
	prURL, err := adapters.EnsurePR(runnerRepo, branch, "main",
		fmt.Sprintf("Shadow: %s vs %s#%s", consumerRepo, workflowsRepo, workflowsPR), body, token)
	if err != nil {
		return err
	}
	fmt.Printf("shadow PR: %s (head %s)\n", prURL, sha)

	return adapters.WatchCommitRun(runnerRepo, sha, token)
}

func main() {
	var workflowsRepo, workflowsRef, consumerRepo, consumerRef, workflowsPR, branch, runnerRepo string
	cmd := &cobra.Command{
		Use:           "mirror-and-test",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := ghactions.RequireFlags([]ghactions.Pair{
				{Key: "workflows-repo", Value: workflowsRepo},
				{Key: "workflows-ref", Value: workflowsRef},
				{Key: "consumer-repo", Value: consumerRepo},
				{Key: "consumer-ref", Value: consumerRef},
				{Key: "workflows-pr", Value: workflowsPR},
				{Key: "branch", Value: branch},
				{Key: "runner-repo", Value: runnerRepo},
			}); err != nil {
				return err
			}
			return run(workflowsRepo, workflowsRef, consumerRepo, consumerRef, workflowsPR, branch, runnerRepo)
		},
	}
	f := cmd.Flags()
	f.StringVar(&workflowsRepo, "workflows-repo", "", "owner/repo of the workflows")
	f.StringVar(&workflowsRef, "workflows-ref", "", "workflows PR head SHA")
	f.StringVar(&consumerRepo, "consumer-repo", "", "owner/repo of the consumer")
	f.StringVar(&consumerRef, "consumer-ref", "", "consumer ref to mirror")
	f.StringVar(&workflowsPR, "workflows-pr", "", "workflows PR number")
	f.StringVar(&branch, "branch", "", "deterministic shadow branch name")
	f.StringVar(&runnerRepo, "runner-repo", "", "owner/repo of the runner")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
