package cmd

import (
	"fmt"

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
PATH signal stay in this list and are omitted from the default audit.

--format=markdown prints the README tables. That output has no install
status and no color. Paste it between the catalog markers in README.md.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stdout := cmd.OutOrStdout()
		// Read the flag here. A package variable would be shared by every
		// test that executes the root command, the same problem the audit
		// flags used to have.
		format, err := cmd.Flags().GetString("format")
		if err != nil {
			return err
		}
		switch format {
		case "text", "":
			output.WriteTitle(stdout)
			// The list is the catalog, so it always walks every registry entry.
			// Installation status is part of that catalog view.
			output.WriteListResults(stdout, check.RunFor(tools.Registry()))
			return nil
		case "markdown":
			// Markdown is the documentation catalog, not a live audit.
			// It has no install status and no color, so it can be pasted
			// between the README markers.
			fmt.Fprint(stdout, tools.CatalogMarkdown(tools.Registry()))
			return nil
		default:
			return fmt.Errorf("unsupported list format %q (supported: text, markdown)", format)
		}
	},
}

func init() {
	listCmd.Flags().String("format", "text", "Output format: text or markdown")
}
