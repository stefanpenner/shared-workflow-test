// Command cleanup (workflows, on pull_request: closed): tear down every consumer's shadow PR +
// branch in the runner for this workflows PR.
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

func main() {
	var runnerRepo, workflowsPR, consumersFile string
	cmd := &cobra.Command{
		Use:           "cleanup",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := ghactions.RequireFlags([]ghactions.Pair{
				{Key: "runner-repo", Value: runnerRepo},
				{Key: "workflows-pr", Value: workflowsPR},
				{Key: "consumers-file", Value: consumersFile},
			}); err != nil {
				return err
			}
			token, err := core.RequireEnv("SHADOW_PAT")
			if err != nil {
				return err
			}
			data, err := os.ReadFile(consumersFile)
			if err != nil {
				return err
			}
			consumers, err := core.ParseConsumers(string(data))
			if err != nil {
				return err
			}
			pr, _ := strconv.Atoi(workflowsPR)
			for _, c := range consumers {
				branch := core.ShadowBranchName(pr, c.Repo)
				if err := adapters.ClosePRAndDeleteBranch(runnerRepo, branch, token); err != nil {
					return err
				}
				fmt.Printf("cleaned up %s:%s\n", runnerRepo, branch)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&runnerRepo, "runner-repo", "", "owner/repo of the runner")
	f.StringVar(&workflowsPR, "workflows-pr", "", "workflows PR number")
	f.StringVar(&consumersFile, "consumers-file", "", "path to shadow-consumers.json")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
