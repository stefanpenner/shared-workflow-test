// Command setup is the Setup action entrypoint: parse named flags, render the section, and write
// node_version to the $GITHUB_OUTPUT sink. Thin — all logic lives in internal/actions/setup.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/stefanpenner-cs/reusable-workflows/internal/actions/setup"
	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
)

func main() {
	var projectName, nodeVersion string

	cmd := &cobra.Command{
		Use:           "setup",
		Short:         "Set up the project environment",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := ghactions.RequireFlags([]ghactions.Pair{
				{Key: "project-name", Value: projectName},
				{Key: "node-version", Value: nodeVersion},
			}); err != nil {
				return err
			}
			version, err := setup.ResolveNodeVersion(nodeVersion)
			if err != nil {
				return err
			}
			fmt.Println(setup.Report(projectName, version))
			return ghactions.AppendOutput(os.Getenv("GITHUB_OUTPUT"), []ghactions.Pair{
				{Key: "node_version", Value: version},
			})
		},
	}
	cmd.Flags().StringVar(&projectName, "project-name", "", "name of the project")
	cmd.Flags().StringVar(&nodeVersion, "node-version", "", "node version to use")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
