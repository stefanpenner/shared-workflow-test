// Command test is the Test action entrypoint: parse named flags and render the section.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	testaction "github.com/stefanpenner-cs/reusable-workflows/internal/actions/test"
	"github.com/stefanpenner-cs/reusable-workflows/internal/ghactions"
)

func main() {
	var suite, coverage string

	cmd := &cobra.Command{
		Use:           "test",
		Short:         "Run the test suite",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := ghactions.RequireFlags([]ghactions.Pair{
				{Key: "suite", Value: suite},
				{Key: "coverage", Value: coverage},
			}); err != nil {
				return err
			}
			fmt.Println(testaction.Report(suite, coverage))
			return nil
		},
	}
	cmd.Flags().StringVar(&suite, "suite", "", "which test suite to run")
	cmd.Flags().StringVar(&coverage, "coverage", "", "enable coverage reporting")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
