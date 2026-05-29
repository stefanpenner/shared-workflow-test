// Command resolve-ctx (workflows setup): resolve the PR number + head SHA to shadow-test and emit
// them on $GITHUB_OUTPUT. Inputs vary by event, so flags are optional; GH_TOKEN (env) authorizes
// the dispatch-time head-SHA lookup.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
	"github.com/stefanpenner-cs/reusable-workflows/internal/shadow/adapters"
	"github.com/stefanpenner-cs/reusable-workflows/internal/shadow/core"
)

func main() {
	var eventName, prNumber, headSha, inputPr, workflowsRepo string

	cmd := &cobra.Command{
		Use:           "resolve-ctx",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx, err := core.ResolveContext(eventName, prNumber, headSha, inputPr, func(pr string) (string, error) {
				if workflowsRepo == "" {
					return "", fmt.Errorf("missing required --workflows-repo for the dispatch lookup")
				}
				return adapters.PullRequestHeadSHA(workflowsRepo, pr, os.Getenv("GH_TOKEN"))
			})
			if err != nil {
				return err
			}
			if err := ghactions.AppendOutput(os.Getenv("GITHUB_OUTPUT"), []ghactions.Pair{
				{Key: "pr", Value: ctx.PR}, {Key: "sha", Value: ctx.SHA},
			}); err != nil {
				return err
			}
			fmt.Printf("✅ resolved PR #%s  https://github.com/%s/pull/%s\n", ctx.PR, workflowsRepo, ctx.PR)
			fmt.Printf("   head %s  %s\n", core.ShortSHA(ctx.SHA), core.CommitURL(workflowsRepo, ctx.SHA))
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&eventName, "event-name", "", "the triggering event name")
	f.StringVar(&prNumber, "pr-number", "", "PR number (pull_request event)")
	f.StringVar(&headSha, "head-sha", "", "head SHA (pull_request event)")
	f.StringVar(&inputPr, "input-pr", "", "PR number (workflow_dispatch input)")
	f.StringVar(&workflowsRepo, "workflows-repo", "", "owner/repo for the dispatch lookup")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
