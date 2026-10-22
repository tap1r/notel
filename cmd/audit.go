package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/stennie/notel/internal/check"
	"github.com/stennie/notel/internal/fix"
	"github.com/stennie/notel/internal/output"
)

// auditOptions is the flag set for one audit invocation.
// It is a value, not package state, so tests can call runAudit without
// mutating the Cobra command's flags.
type auditOptions struct {
	ShowAll bool
	Verbose bool
	Fix     string
	Profile bool
}

// auditRuntime is the process boundary behind runAudit.
// Production uses the real checker, shell detector, and renderer.
// Tests pass a value with the pieces they care about replaced. Nothing
// in this package is a mutable function variable.
type auditRuntime struct {
	run         func() []check.ToolResult
	detectShell func() (fix.Shell, error)
	parseShell  func(string) (fix.Shell, error)
	buildPlan   func([]check.ToolResult, bool, fix.Mode) fix.Plan
	title       func(io.Writer)
	compact     func(io.Writer, []check.ToolResult, bool)
	detailed    func(io.Writer, []check.ToolResult, bool)
	summary     func(io.Writer, []check.ToolResult)
	fixes       func(io.Writer, io.Writer, fix.Shell, fix.Plan)
}

func productionAuditRuntime() auditRuntime {
	return auditRuntime{
		run:         check.Run,
		detectShell: fix.DetectShell,
		parseShell:  fix.ParseShell,
		buildPlan:   fix.Build,
		title:       output.WriteTitle,
		compact:     output.WriteCheckResults,
		detailed:    output.WriteReportResults,
		summary:     output.WriteSummary,
		fixes:       output.PrintFixCommands,
	}
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit telemetry settings for installed tools",
	Args:  cobra.NoArgs,
	Long: `Checks all installed developer tools and reports whether their
telemetry opt-out environment variables are correctly set.

By default only installed tools are shown. Use --all to include tools
that are not detected in PATH. Use --verbose for detailed per-tool output.
Use --fix[=<shell>] to print the shell commands that failing tools still need.
Use --fix --profile to print every visible opt-out, including values that already match.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := auditOptionsFrom(cmd)
		if err != nil {
			return err
		}
		return runAudit(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts, productionAuditRuntime())
	},
}

func init() {
	auditCmd.Flags().BoolP("all", "a", false, "Show all tools, including those not installed")
	auditCmd.Flags().BoolP("verbose", "v", false, "Show detailed per-tool output")
	auditCmd.Flags().StringP("fix", "f", "", "Print shell commands to disable telemetry; optionally pass bash, fish, powershell, or zsh")
	auditCmd.Flags().Lookup("fix").NoOptDefVal = "auto"
	auditCmd.Flags().Bool("profile", false, "With --fix, print every visible opt-out, including values that already match")
}

func auditOptionsFrom(cmd *cobra.Command) (auditOptions, error) {
	showAll, err := cmd.Flags().GetBool("all")
	if err != nil {
		return auditOptions{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return auditOptions{}, err
	}
	fixValue, err := cmd.Flags().GetString("fix")
	if err != nil {
		return auditOptions{}, err
	}
	profile, err := cmd.Flags().GetBool("profile")
	if err != nil {
		return auditOptions{}, err
	}
	return auditOptions{
		ShowAll: showAll,
		Verbose: verbose,
		Fix:     fixValue,
		Profile: profile,
	}, nil
}

// runAudit is the audit command without Cobra. stdout receives the report,
// or the shell script when Fix is set. stderr receives the report alongside
// a fix script, plus errors about the shell. The two streams stay separate
// so a redirect of stdout captures only the script.
func runAudit(stdout, stderr io.Writer, opts auditOptions, rt auditRuntime) error {
	results := rt.run()
	if opts.Fix != "" {
		shell, err := resolveFixShell(opts.Fix, rt)
		if err != nil {
			return err
		}
		rt.title(stderr)
		if opts.ShowAll || opts.Verbose {
			writeAuditView(stderr, results, opts, rt)
		}
		mode := fix.ModeDelta
		if opts.Profile {
			// --profile keeps every visible opt-out, including values that
			// already match. The default --fix prints only what failing tools
			// still need.
			mode = fix.ModeProfile
		}
		rt.fixes(stdout, stderr, shell, rt.buildPlan(results, opts.ShowAll, mode))
		return nil
	}
	if opts.Profile {
		return fmt.Errorf("--profile is used together with --fix")
	}

	rt.title(stdout)
	writeAuditView(stdout, results, opts, rt)
	return nil
}

func resolveFixShell(value string, rt auditRuntime) (fix.Shell, error) {
	if value == "" || value == "auto" {
		shell, err := rt.detectShell()
		if err != nil {
			return "", fmt.Errorf("%w; use --fix=<shell> to choose one explicitly", err)
		}
		return shell, nil
	}
	return rt.parseShell(value)
}

func writeAuditView(w io.Writer, results []check.ToolResult, opts auditOptions, rt auditRuntime) {
	if opts.Verbose {
		rt.detailed(w, results, opts.ShowAll)
	} else {
		rt.compact(w, results, opts.ShowAll)
	}
	rt.summary(w, results)
}
