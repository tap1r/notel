// Package output provides lipgloss-styled rendering for notel's check results.
package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stennie/notel/internal/check"
	"github.com/stennie/notel/internal/fix"
	"github.com/stennie/notel/internal/tools"
)

// ── Palette ─────────────────────────────────────────────────────────────────

var (
	colorGreen  = lipgloss.Color("2")
	colorRed    = lipgloss.Color("1")
	colorYellow = lipgloss.Color("3")
	colorCyan   = lipgloss.Color("6")
	colorPurple = lipgloss.Color("99")
	colorBlue   = lipgloss.Color("69")
	colorGray   = lipgloss.Color("240")
	colorDim    = lipgloss.Color("8")
)

// ── Styles ───────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPurple)

	categoryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue).
			MarginTop(1)

	sectionDivider = lipgloss.NewStyle().
			Foreground(colorGray).
			Render(strings.Repeat("─", 60))

	passingIcon = lipgloss.NewStyle().Foreground(colorGreen).Render("✓")
	failingIcon = lipgloss.NewStyle().Foreground(colorRed).Render("✗")
	skippedIcon = lipgloss.NewStyle().Foreground(colorDim).Render("–")
	warningIcon = lipgloss.NewStyle().Foreground(colorYellow).Render("!")

	passingStyle = lipgloss.NewStyle().Foreground(colorGreen)
	failingStyle = lipgloss.NewStyle().Foreground(colorRed)
	skippedStyle = lipgloss.NewStyle().Foreground(colorDim)
	boldStyle    = lipgloss.NewStyle().Bold(true)
	faintStyle   = lipgloss.NewStyle().Faint(true)
	envStyle     = lipgloss.NewStyle().Foreground(colorYellow).Italic(true)
	valueStyle   = lipgloss.NewStyle().Foreground(colorGreen)
	hintStyle    = lipgloss.NewStyle().Foreground(colorCyan).Faint(true)

	summaryBoxStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 2).
			MarginTop(1)

	helpHeadingStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorBlue).
				MarginTop(1)

	commandNameStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPurple)

	flagNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorYellow)
)

// ── Public API ───────────────────────────────────────────────────────────────

// WriteTitle renders the notel application header to a specific writer.
func WriteTitle(w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, titleStyle.Render("  notel")+faintStyle.Render("  —  DevTools Telemetry Auditor"))
	fmt.Fprintln(w, faintStyle.Render("  "+sectionDivider))
}

// PrintHelp renders styled Cobra help output consistent with the rest of the CLI.
func PrintHelp(w io.Writer, cmd *cobra.Command, copyright string) {
	WriteTitle(w)

	description := cmd.Long
	if description == "" {
		description = cmd.Short
	}
	if description != "" {
		fmt.Fprintln(w, wrapHelpText(description))
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, helpHeadingStyle.Render("Usage"))
	// Pad the plain text, then style it. Padding an already-styled string
	// counts ANSI escapes as columns, so the help grid drifts whenever color is on.
	fmt.Fprintf(w, "  %s\n", commandNameStyle.Render(cmd.UseLine()))

	commands := visibleCommands(cmd)
	if len(commands) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, helpHeadingStyle.Render("Available Commands"))
		for _, sub := range commands {
			fmt.Fprintf(w, "  %s %s\n",
				commandNameStyle.Render(pad(sub.Name(), 16)),
				faintStyle.Render(sub.Short),
			)
		}
	}

	writeFlagSection(w, "Flags", cmd.LocalFlags())
	writeFlagSection(w, "Global Flags", cmd.InheritedFlags())

	if cmd.HasAvailableSubCommands() {
		fmt.Fprintln(w)
		fmt.Fprintln(w, hintStyle.Render(fmt.Sprintf(`Use "%s [command] --help" for more information about a command.`, cmd.CommandPath())))
	}

	if copyright != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, faintStyle.Render(copyright))
	}
}

// PrintFixCommands renders scriptable shell commands to stdout and optional context to stderr.
//
// stdout is only the script, so it can be redirected into a shell rc file.
// The reason a script is empty, and any disagreement about a shared variable,
// is written to stderr. An empty plan used to claim that telemetry was already
// disabled, including when the reason was that no tools were installed.
func PrintFixCommands(stdout io.Writer, stderr io.Writer, shell fix.Shell, plan fix.Plan) {
	if stderr != nil {
		fmt.Fprintln(stderr, helpHeadingStyle.Render("Fix Commands"))
		fmt.Fprintf(stderr, "  %s\n", faintStyle.Render("Shell: "+string(shell)))
		switch plan.Status {
		case fix.StatusNoneInstalled:
			fmt.Fprintln(stderr)
			fmt.Fprintln(stderr, warningIcon+"  "+lipgloss.NewStyle().Foreground(colorYellow).Render(
				"No supported tools detected in PATH; no fix commands to print.",
			))
			fmt.Fprintln(stderr, hintStyle.Render("  Re-run with --all to print commands for tools that are not installed."))
			return
		case fix.StatusAlreadyDisabled:
			fmt.Fprintln(stderr)
			fmt.Fprintln(stderr, passingStyle.Render("  All visible tools already have telemetry disabled; no fix commands needed."))
			return
		}
		if plan.Mode == fix.ModeProfile {
			fmt.Fprintln(stderr, hintStyle.Render("  Profile: every visible opt-out, including values that already match."))
		}
		for _, conflict := range plan.Conflicts {
			fmt.Fprintf(stderr, "  %s %s is %s from %s; %s wants %s. Keeping %s.\n",
				warningIcon,
				conflict.Name,
				conflict.Kept,
				conflict.KeptTool,
				conflict.OtherTool,
				conflict.Other,
				conflict.Kept,
			)
		}
	}

	for i, block := range scriptBlocks(shell, plan) {
		if i > 0 {
			fmt.Fprintln(stdout)
		}
		if block.category != "" {
			fmt.Fprintln(stdout, fix.Comment(block.category))
		}
		for _, line := range block.lines {
			fmt.Fprintln(stdout, line)
		}
	}
}

type scriptBlock struct {
	category string
	lines    []string
}

// scriptBlocks groups export lines and unverified CLI comments by category.
//
// Suggestion order is the order tools were audited, which is registry order.
// A shared variable is already deduplicated inside the plan, so DO_NOT_TRACK
// appears in the category of the first tool that contributed it. CLI hints
// are shell comments. The same command text is printed once.
func scriptBlocks(shell fix.Shell, plan fix.Plan) []scriptBlock {
	var blocks []scriptBlock
	index := map[string]int{}
	add := func(category, line string) {
		i, ok := index[category]
		if !ok {
			index[category] = len(blocks)
			blocks = append(blocks, scriptBlock{category: category})
			i = len(blocks) - 1
		}
		blocks[i].lines = append(blocks[i].lines, line)
	}

	for _, suggestion := range plan.Suggestions {
		add(suggestion.Category, fix.Command(shell, suggestion))
	}
	seenCommand := map[string]bool{}
	for _, hint := range plan.CommandHints {
		if seenCommand[hint.Command] {
			continue
		}
		seenCommand[hint.Command] = true
		// The leading "# " keeps the hint out of the shell's execution if the
		// user pastes the whole script. The audit does not know whether the
		// command has already been run.
		add(hint.Category, "# "+hint.Command)
	}
	return blocks
}

func visibleCommands(cmd *cobra.Command) []*cobra.Command {
	commands := make([]*cobra.Command, 0, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && !sub.IsAdditionalHelpTopicCommand() {
			commands = append(commands, sub)
		}
	}
	return commands
}

func writeFlagSection(w io.Writer, heading string, flags *pflag.FlagSet) {
	if flags == nil || !flags.HasAvailableFlags() {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, helpHeadingStyle.Render(heading))
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		var names []string
		if flag.Shorthand != "" {
			names = append(names, "-"+flag.Shorthand)
		}
		names = append(names, "--"+flag.Name)

		label := strings.Join(names, ", ")
		if flag.Value.Type() != "bool" {
			label += " <value>"
		}
		usage := flag.Usage
		if flag.DefValue != "" && flag.DefValue != "false" && flag.DefValue != "[]" {
			// The default stays in the usage column. Putting it inside the
			// padded name counted the styled default toward the column width.
			usage += " (default: " + flag.DefValue + ")"
		}

		fmt.Fprintf(w, "  %s %s\n",
			flagNameStyle.Render(pad(label, 22)),
			faintStyle.Render(usage),
		)
	})
}

// pad appends spaces so the visible width is at least width.
// Call it on plain text. Styling afterwards does not change how many
// columns the terminal advances.
func pad(text string, width int) string {
	runes := []rune(text)
	if len(runes) >= width {
		return text
	}
	return text + strings.Repeat(" ", width-len(runes))
}

func wrapHelpText(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// WriteCheckResults renders a compact per-tool status list grouped by category.
// When showAll is false, uninstalled tools are omitted.
func WriteCheckResults(w io.Writer, results []check.ToolResult, showAll bool) {
	categories, order := groupByCategory(results)

	for _, cat := range order {
		items := categories[cat]
		visible := filterInstalled(items, showAll)
		if len(visible) == 0 {
			continue
		}

		fmt.Fprintln(w, categoryStyle.Render("  "+cat))
		for _, r := range visible {
			writeCompactRow(w, r)
		}
	}
}

// WriteReportResults renders a detailed per-tool report grouped by category.
// When showAll is false, uninstalled tools are omitted.
func WriteReportResults(w io.Writer, results []check.ToolResult, showAll bool) {
	categories, order := groupByCategory(results)

	for _, cat := range order {
		visible := filterInstalled(categories[cat], showAll)
		if len(visible) == 0 {
			continue
		}

		fmt.Fprintln(w, categoryStyle.Render("  "+cat))
		fmt.Fprintln(w, "  "+sectionDivider)

		for _, r := range visible {
			writeDetailedBlock(w, r)
		}
		fmt.Fprintln(w)
	}
}

// WriteListResults renders the full registry with env-var documentation
// and the installation status the audit just computed.
func WriteListResults(w io.Writer, results []check.ToolResult) {
	fmt.Fprintln(w, boldStyle.Render("  Supported Tools"))
	fmt.Fprintln(w)
	categories, order := groupByCategory(results)

	for _, cat := range order {
		fmt.Fprintln(w, categoryStyle.Render("  "+cat))

		for _, r := range categories[cat] {
			fmt.Fprintf(w, "  %s %s\n", boldStyle.Render(pad(r.Tool.Name, 22)), faintStyle.Render(r.Tool.Description))
			fmt.Fprintf(w, "    %s %s\n", faintStyle.Render("Status:"), hintStyle.Render(installStatus(r)))
			fmt.Fprintf(w, "    %s %s\n", faintStyle.Render("Data:"), hintStyle.Render(r.Tool.DataCollection))
			fmt.Fprintf(w, "    %s %s\n", faintStyle.Render("Docs:"), hintStyle.Render(r.Tool.DocumentationURL))

			if r.Tool.OptOut == tools.AnyOf {
				fmt.Fprintf(w, "    %s\n", hintStyle.Render("Any one of these variables opts out."))
			}
			if r.Tool.OptOut == tools.FirstSet {
				fmt.Fprintf(w, "    %s\n", hintStyle.Render("The first variable that is set decides. Later variables apply only when it is absent."))
			}
			if r.Tool.Default == tools.DefaultDisabled {
				fmt.Fprintf(w, "    %s\n", hintStyle.Render("Unset means already off. Collection is opt-in."))
			}
			for _, c := range r.Tool.EnvChecks {
				validStr := formatValidValues(c.ValidValues)
				if validStr == "" {
					validStr = "<any non-empty value>"
				}
				if c.IgnoreCase {
					validStr += " (any case)"
				}
				fmt.Fprintf(w, "    %s %s\n",
					envStyle.Render(c.Name),
					faintStyle.Render("= "+validStr),
				)
				fmt.Fprintf(w, "    %s\n", hintStyle.Render(c.Description))
			}
			for _, command := range r.Tool.Commands {
				fmt.Fprintf(w, "    %s %s\n", faintStyle.Render("CLI:"), hintStyle.Render(command))
			}
			fmt.Fprintln(w)
		}
	}
}

// WriteSummary renders the aggregate counts at the bottom of a check/report run.
func WriteSummary(w io.Writer, results []check.ToolResult) {
	s := check.Summarise(results)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "  "+sectionDivider)

	var msg string
	switch {
	case s.Installed == 0:
		msg = warningIcon + "  " + lipgloss.NewStyle().Foreground(colorYellow).Render(
			"No supported tools detected in PATH",
		)
	case s.Failing == 0:
		msg = passingIcon + "  " + passingStyle.Render(
			fmt.Sprintf("All %d installed tools have telemetry disabled", s.Installed),
		)
	default:
		msg = fmt.Sprintf("%s  %s  %s",
			failingIcon,
			boldStyle.Render(fmt.Sprintf("%d/%d", s.Passing, s.Installed)),
			"installed tools have telemetry disabled  "+
				failingStyle.Render(fmt.Sprintf("(%d need attention)", s.Failing)),
		)
	}

	fmt.Fprintln(w, summaryBoxStyle.Render(msg))
	fmt.Fprintln(w)
}

// ── Internal helpers ─────────────────────────────────────────────────────────

func writeCompactRow(w io.Writer, r check.ToolResult) {
	const nameWidth = 20

	if !r.Installed {
		name := skippedStyle.Render(fmt.Sprintf("%-*s", nameWidth, r.Tool.Name))
		envParts := detectedEnvParts(r.Checks)
		if len(envParts) == 0 {
			fmt.Fprintf(w, "  %s  %s  %s\n", skippedIcon, name, faintStyle.Render("not installed"))
			return
		}
		fmt.Fprintf(w, "  %s  %s  %s  %s\n", skippedIcon, name, faintStyle.Render("not installed"), strings.Join(envParts, "  "))
		return
	}

	icon := passingIcon
	name := boldStyle.Render(fmt.Sprintf("%-*s", nameWidth, r.Tool.Name))
	if !r.AllPassing {
		icon = failingIcon
	}

	var envParts []string
	for i, c := range r.Checks {
		switch {
		case c.Reason == check.ReasonDefaultOff:
			// Unset is a pass for opt-in tools. Keep it visually distinct from
			// a variable the user explicitly set to an opt-out value.
			envParts = append(envParts,
				faintStyle.Render(c.Check.Name)+" "+hintStyle.Render("(not set; off unless you opt in)"),
			)
		case ignoredBecauseEarlierVariableIsSet(r, i):
			// GH_TELEMETRY=1 with DO_NOT_TRACK=1 still uploads. The later
			// match must not be painted as the reason the row failed.
			envParts = append(envParts,
				faintStyle.Render(c.Check.Name)+" "+hintStyle.Render("(not consulted; an earlier variable is set)"),
			)
		case c.Passing:
			envParts = append(envParts,
				passingStyle.Render(c.Check.Name)+"="+valueStyle.Render(c.Value),
			)
		case unusedBecauseTheToolPassed(r, c):
			// Another check already satisfied the tool. Show this one as
			// unused instead of as a failure, or the row would be green and red
			// at the same time.
			envParts = append(envParts,
				faintStyle.Render(c.Check.Name)+" "+hintStyle.Render("("+unusedCheckNote(r, i)+")"),
			)
		case c.Value == "":
			envParts = append(envParts,
				failingStyle.Render(c.Check.Name)+" "+faintStyle.Render("(not set)"),
			)
		default:
			envParts = append(envParts,
				failingStyle.Render(c.Check.Name+"="+c.Value)+" "+
					hintStyle.Render("(unexpected value; assuming telemetry enabled)"),
			)
		}
	}

	fmt.Fprintf(w, "  %s  %s  %s\n", icon, name, strings.Join(envParts, "  "))
}

func writeDetailedBlock(w io.Writer, r check.ToolResult) {
	icon := skippedIcon
	if r.Installed {
		if r.AllPassing {
			icon = passingIcon
		} else {
			icon = failingIcon
		}
	}

	installed := faintStyle.Render("(not installed)")
	if r.Installed {
		installed = faintStyle.Render("(" + r.BinaryPath + ")")
	}

	fmt.Fprintf(w, "\n  %s  %s  %s\n",
		icon,
		boldStyle.Render(r.Tool.Name),
		installed,
	)
	fmt.Fprintf(w, "       %s %s\n", faintStyle.Render("Docs:"), hintStyle.Render(r.Tool.DocumentationURL))
	if r.Tool.OptOut == tools.AnyOf {
		fmt.Fprintf(w, "       %s\n", hintStyle.Render("Any one of these variables opts out."))
	}
	if r.Tool.OptOut == tools.FirstSet {
		fmt.Fprintf(w, "       %s\n", hintStyle.Render("The first variable that is set decides. Later variables apply only when it is absent."))
	}

	if !r.Installed {
		// Uninstalled tools stay quiet unless the environment already has a
		// relevant value, or the entry has a CLI hint worth showing under --all.
		if !hasDetectedEnvValue(r.Checks) && len(r.Tool.Commands) == 0 {
			return
		}
		for i, c := range r.Checks {
			if c.Value == "" && c.Reason != check.ReasonDefaultOff {
				continue
			}
			fmt.Fprintln(w, detailedStatusLine(r, i))
			fmt.Fprintf(w, "       %s\n", faintStyle.Render(c.Check.Description))
		}
		writeCommandHints(w, r.Tool)
		return
	}

	for i, c := range r.Checks {
		fmt.Fprintln(w, detailedStatusLine(r, i))
		fmt.Fprintf(w, "       %s\n", faintStyle.Render(c.Check.Description))
	}
	writeCommandHints(w, r.Tool)
}

// writeCommandHints prints CLI remediations that notel does not verify.
// They sit next to the env checks so a passing crash-reporting variable
// cannot hide a Flutter analytics setting that has no environment variable.
func writeCommandHints(w io.Writer, tool tools.Tool) {
	for _, command := range tool.Commands {
		fmt.Fprintf(w, "       %s %s\n",
			hintStyle.Render("→ also (not verified):"),
			hintStyle.Render(command),
		)
	}
}

func detailedStatusLine(r check.ToolResult, index int) string {
	c := r.Checks[index]
	validStr := formatValidValues(c.Check.ValidValues)
	validStr = strings.ReplaceAll(validStr, " | ", " or ")
	if c.Check.IgnoreCase && validStr != "" {
		validStr += " (any case)"
	}
	if c.Reason == check.ReasonDefaultOff {
		return fmt.Sprintf("     %s  %s  %s",
			passingStyle.Render("✓"),
			envStyle.Render(c.Check.Name),
			hintStyle.Render("not set; telemetry is off unless you opt in"),
		)
	}
	if ignoredBecauseEarlierVariableIsSet(r, index) {
		return fmt.Sprintf("     %s  %s  %s",
			skippedStyle.Render("–"),
			envStyle.Render(c.Check.Name),
			hintStyle.Render("not consulted; an earlier variable is set"),
		)
	}
	if c.Passing {
		return fmt.Sprintf("     %s  %s=%s",
			passingStyle.Render("✓"),
			envStyle.Render(c.Check.Name),
			valueStyle.Render(c.Value),
		)
	}
	if unusedBecauseTheToolPassed(r, c) {
		return fmt.Sprintf("     %s  %s  %s",
			skippedStyle.Render("–"),
			envStyle.Render(c.Check.Name),
			hintStyle.Render(unusedCheckNote(r, index)),
		)
	}
	if c.Value == "" {
		statusLine := fmt.Sprintf("     %s  %s  %s",
			failingStyle.Render("✗"),
			envStyle.Render(c.Check.Name),
			failingStyle.Render("not set"),
		)
		return statusLine + fmt.Sprintf("\n       %s %s=%s",
			hintStyle.Render("→ set"),
			envStyle.Render(c.Check.Name),
			hintStyle.Render(validStr),
		)
	}
	statusLine := fmt.Sprintf("     %s  %s=%s  %s",
		warningIcon,
		envStyle.Render(c.Check.Name),
		failingStyle.Render(c.Value),
		hintStyle.Render("(unexpected value; assuming telemetry enabled)"),
	)
	return statusLine + fmt.Sprintf("\n       %s %s=%s",
		hintStyle.Render("→ set"),
		envStyle.Render(c.Check.Name),
		hintStyle.Render(validStr),
	)
}

func formatValidValues(values []string) string {
	parts := make([]string, len(values))
	for i, value := range values {
		if value == "" {
			parts[i] = "(empty)"
			continue
		}
		parts[i] = value
	}
	return strings.Join(parts, " | ")
}

func firstPresentCheck(r check.ToolResult) int {
	for i, c := range r.Checks {
		if c.Reason != check.ReasonUnset && c.Reason != check.ReasonDefaultOff {
			return i
		}
	}
	return -1
}

func ignoredBecauseEarlierVariableIsSet(r check.ToolResult, index int) bool {
	if r.Tool.OptOut != tools.FirstSet || r.AllPassing {
		return false
	}
	decisive := firstPresentCheck(r)
	return decisive >= 0 && index > decisive
}

func unusedBecauseTheToolPassed(r check.ToolResult, c check.EnvCheckResult) bool {
	if !r.AllPassing || c.Passing {
		return false
	}
	return r.Tool.OptOut == tools.AnyOf || r.Tool.OptOut == tools.FirstSet
}

func unusedCheckNote(r check.ToolResult, index int) string {
	if r.Tool.OptOut == tools.AnyOf {
		return "alternative opt-out, not required"
	}
	if decisive := firstPresentCheck(r); decisive >= 0 && index > decisive {
		return "not consulted; an earlier variable decided"
	}
	return "not set; a later variable decided"
}

// installStatus is the list-view label for how the tool was detected.
//
// AlwaysInstalled conventions have no path. An empty Detect list that is
// not always installed has no PATH signal we trust, which is different
// from a binary that was looked up and not found.
func installStatus(r check.ToolResult) string {
	switch {
	case r.Tool.AlwaysInstalled:
		return "env-only (no binary)"
	case r.Installed && r.BinaryPath != "":
		return "installed (" + r.BinaryPath + ")"
	case r.Installed:
		return "installed"
	case len(r.Tool.Detect) == 0:
		return "not detected (no PATH signal)"
	default:
		return "not installed"
	}
}

func hasDetectedEnvValue(checks []check.EnvCheckResult) bool {
	for _, c := range checks {
		if c.Value != "" {
			return true
		}
	}
	return false
}

func detectedEnvParts(checks []check.EnvCheckResult) []string {
	parts := make([]string, 0, len(checks))
	for _, c := range checks {
		if c.Value == "" {
			continue
		}
		if c.Passing {
			parts = append(parts,
				passingStyle.Render(c.Check.Name)+"="+valueStyle.Render(c.Value),
			)
			continue
		}
		parts = append(parts,
			failingStyle.Render(c.Check.Name+"="+c.Value)+" "+
				hintStyle.Render("(unexpected value; assuming telemetry enabled)"),
		)
	}
	return parts
}

func filterInstalled(results []check.ToolResult, showAll bool) []check.ToolResult {
	if showAll {
		return results
	}
	out := make([]check.ToolResult, 0, len(results))
	for _, r := range results {
		if r.Installed {
			out = append(out, r)
		}
	}
	return out
}

func groupByCategory(results []check.ToolResult) (map[string][]check.ToolResult, []string) {
	groups := make(map[string][]check.ToolResult)
	var order []string
	seen := make(map[string]bool)

	for _, r := range results {
		cat := r.Tool.Category
		if !seen[cat] {
			seen[cat] = true
			order = append(order, cat)
		}
		groups[cat] = append(groups[cat], r)
	}
	return groups, order
}
