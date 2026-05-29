// Command dispatch-and-watch (workflows, one invocation per consumer): dispatch the runner
// receiver, watch it to completion, and render the result into the job summary + logs. The job's
// exit status is the PR's `Shadow: <consumer>` check.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
	"github.com/stefanpenner-cs/reusable-workflows/internal/shadow/adapters"
	"github.com/stefanpenner-cs/reusable-workflows/internal/shadow/core"
)

func run(runnerRepo, workflowsRepo, workflowsRefArg, workflowsPRArg, consumerRepo, consumerRef string) error {
	token, err := core.RequireEnv("SHADOW_PAT")
	if err != nil {
		return err
	}
	workflowsPR, _ := strconv.Atoi(workflowsPRArg)
	branch := core.ShadowBranchName(workflowsPR, consumerRepo)
	ctx := core.ShadowContext{
		WorkflowsRepo: workflowsRepo,
		WorkflowsRef:  workflowsRefArg,
		ConsumerRepo:  consumerRepo,
		ConsumerRef:   consumerRef,
		WorkflowsPR:   workflowsPR,
		Branch:        branch,
	}

	runID, err := adapters.DispatchReceiver(runnerRepo, ctx, token)
	if err != nil {
		return err
	}
	runURL := fmt.Sprintf("https://github.com/%s/actions/runs/%d", runnerRepo, runID)
	fmt.Printf("🛰️  Shadow test: %s@%s\n", consumerRepo, consumerRef)
	fmt.Printf("    vs %s — runner run: %s\n", core.WorkflowsPrURL(workflowsRepo, workflowsPR), runURL)
	fmt.Printf("::notice title=Shadow test::🛰️ %s — runner run: %s\n", consumerRepo, runURL)

	finish := func(result core.ShadowResult) string {
		prURL, _ := adapters.FindPrURL(runnerRepo, branch, token)
		in := core.ShadowSummaryInput{
			ConsumerRepo: consumerRepo, ConsumerRef: consumerRef,
			WorkflowsRepo: workflowsRepo, WorkflowsRef: workflowsRefArg, WorkflowsPR: workflowsPR,
			Result: result, RunURL: runURL, PRURL: prURL,
		}
		if summary := os.Getenv("GITHUB_STEP_SUMMARY"); summary != "" {
			_ = ghactions.AppendFile(summary, core.RenderShadowSummary(in))
		}
		for _, line := range core.RenderShadowLog(in) {
			fmt.Println(line)
		}
		return prURL
	}

	if err := adapters.WatchRun(runnerRepo, runID, token); err != nil {
		prURL := finish(core.ShadowFailed)
		target := prURL
		if target == "" {
			target = runURL
		}
		fmt.Printf("::error title=Shadow test failed::❌ %s — open %s to see the failing job\n", consumerRepo, target)
		return err
	}
	finish(core.ShadowPassed)
	return nil
}

func main() {
	var runnerRepo, workflowsRepo, workflowsRef, workflowsPR, consumerRepo, consumerRef string
	cmd := &cobra.Command{
		Use:           "dispatch-and-watch",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := ghactions.RequireFlags([]ghactions.Pair{
				{Key: "runner-repo", Value: runnerRepo},
				{Key: "workflows-repo", Value: workflowsRepo},
				{Key: "workflows-ref", Value: workflowsRef},
				{Key: "workflows-pr", Value: workflowsPR},
				{Key: "consumer-repo", Value: consumerRepo},
				{Key: "consumer-ref", Value: consumerRef},
			}); err != nil {
				return err
			}
			return run(runnerRepo, workflowsRepo, workflowsRef, workflowsPR, consumerRepo, consumerRef)
		},
	}
	f := cmd.Flags()
	f.StringVar(&runnerRepo, "runner-repo", "", "owner/repo of the runner")
	f.StringVar(&workflowsRepo, "workflows-repo", "", "owner/repo of the workflows")
	f.StringVar(&workflowsRef, "workflows-ref", "", "workflows PR head SHA")
	f.StringVar(&workflowsPR, "workflows-pr", "", "workflows PR number")
	f.StringVar(&consumerRepo, "consumer-repo", "", "owner/repo of the consumer")
	f.StringVar(&consumerRef, "consumer-ref", "", "consumer ref to mirror")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
