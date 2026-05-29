// Command check-dispatch-auth (security regression test): prove that unauthorized callers cannot
// trigger the runner's receiver via workflow_dispatch. It POSTs the dispatch endpoint with (1) no
// credential and (2) an invalid credential, and asserts GitHub refuses both (401/403/404). It never
// sends a valid token, so it cannot trigger a real shadow run — safe to run anywhere, including in
// Actions where a GITHUB_TOKEN is present (which it ignores). Exits non-zero if any unauthorized
// attempt is NOT refused. See shadow/SECURITY.md for the trust model this guards.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
	"github.com/stefanpenner-cs/reusable-workflows/internal/shadow/adapters"
	"github.com/stefanpenner-cs/reusable-workflows/internal/shadow/core"
)

func run(runnerRepo string) error {
	specs := []struct {
		name string
		auth string
	}{
		{"unauthenticated", ""},
		{"invalid-token", "Bearer ghp_invalidtokenAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
	}

	fmt.Printf("Probing workflow_dispatch auth on %s (receiver.yaml):\n", runnerRepo)
	var probes []core.DispatchProbe
	for _, s := range specs {
		status, err := adapters.ProbeDispatchStatus(runnerRepo, s.auth)
		if err != nil {
			return fmt.Errorf("probe %q: %w", s.name, err)
		}
		probes = append(probes, core.DispatchProbe{Name: s.name, Status: status})
		fmt.Printf("  %-16s → HTTP %d (%s)\n", s.name, status, verdict(status))
	}

	if leaked := core.LeakedProbes(probes); len(leaked) > 0 {
		return fmt.Errorf("dispatch-auth check FAILED: %d unauthorized attempt(s) were not refused: %v", len(leaked), leaked)
	}
	fmt.Printf("✅ all %d unauthorized dispatch attempts were refused\n", len(probes))
	return nil
}

func verdict(status int) string {
	if core.DispatchRejected(status) {
		return "refused"
	}
	return "NOT REFUSED — SECURITY REGRESSION"
}

func main() {
	var runnerRepo string
	cmd := &cobra.Command{
		Use:           "check-dispatch-auth",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := ghactions.RequireFlags([]ghactions.Pair{
				{Key: "runner-repo", Value: runnerRepo},
			}); err != nil {
				return err
			}
			return run(runnerRepo)
		},
	}
	cmd.Flags().StringVar(&runnerRepo, "runner-repo", "", "owner/repo of the runner (shadow-testing) repo")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
