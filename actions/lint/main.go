// Command lint is the Lint action entrypoint: parse named flags and render the section.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/stefanpenner-cs/reusable-workflows/internal/actions/lint"
	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
)

func main() {
	var paths, config string

	cmd := &cobra.Command{
		Use:           "lint",
		Short:         "Run linting checks",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := ghactions.RequireFlags([]ghactions.Pair{
				{Key: "paths", Value: paths},
				{Key: "config", Value: config},
			}); err != nil {
				return err
			}
			fmt.Println(lint.Report(paths, config))
			return nil
		},
	}
	cmd.Flags().StringVar(&paths, "paths", "", "paths to lint")
	cmd.Flags().StringVar(&config, "config", "", "lint config file")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
