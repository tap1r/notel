package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stennie/notel/internal/check"
	"github.com/stennie/notel/internal/output"
	"github.com/stennie/notel/internal/tools"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all supported tools and their telemetry variables",
	Args:    cobra.NoArgs,
	Long: `Lists every developer tool that notel can audit, along with the
environment variables and accepted values that disable telemetry.

Each row includes whether the tool was found on PATH. Tools with no trusted
PATH signal stay in this list and are omitted from the default audit.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stdout := cmd.OutOrStdout()
		output.WriteTitle(stdout)
		// The list is the catalog, so it always walks every registry entry.
		// Installation status is part of that catalog view.
		output.WriteListResults(stdout, check.RunFor(tools.Registry()))
		return nil
	},
}
