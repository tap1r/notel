// Package fix builds shell-specific env-var commands to disable telemetry.
package fix

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stennie/notel/internal/check"
	"github.com/stennie/notel/internal/tools"
)

// Shell identifies a supported shell syntax for fix commands.
type Shell string

const (
	Bash       Shell = "bash"
	Zsh        Shell = "zsh"
	Fish       Shell = "fish"
	PowerShell Shell = "powershell"
)

// Mode selects which opt-outs a fix plan contains.
type Mode int

const (
	// ModeDelta lists what is still required to make failing visible tools pass.
	// Variables that already match are omitted. AnyOf tools contribute one
	// suggested variable, because one alternative is enough.
	ModeDelta Mode = iota

	// ModeProfile lists every visible opt-out, including values that already
	// match. The result is a reusable shell block rather than a minimal patch.
	ModeProfile
)

// Status tells the printer why a plan does or does not contain commands.
// An empty suggestion list is not one situation. It is either "nothing is
// installed" or "everything visible already passes".
type Status int

const (
	// StatusFixes means stdout should contain shell commands.
	StatusFixes Status = iota
	// StatusNoneInstalled means the default audit saw no installed tools.
	// --all is the way to print commands for tools that are not on PATH.
	StatusNoneInstalled
	// StatusAlreadyDisabled means every visible tool already satisfies its policy.
	StatusAlreadyDisabled
)

// Environment reads process variables for shell detection.
// A nil Getenv uses os.Getenv. Tests pass a map-backed function instead of
// replacing a package variable, so detection tests can run in parallel.
type Environment struct {
	Getenv func(string) string
}

func (e Environment) getenv(key string) string {
	if e.Getenv == nil {
		return os.Getenv(key)
	}
	return e.Getenv(key)
}

// Suggestion represents one env var that should be set to disable telemetry.
type Suggestion struct {
	Category    string
	ToolName    string
	Installed   bool
	Current     string
	Check       tools.EnvCheck
	TargetValue string
}

// CommandHint is a CLI remediation printed as a shell comment.
// notel does not run it and does not know whether it has already been run.
type CommandHint struct {
	Category string
	ToolName string
	Command  string
}

// Conflict records two visible tools that want the same variable set to
// different values. The plan keeps the first value and reports the rest.
type Conflict struct {
	Name      string
	Kept      string
	KeptTool  string
	Other     string
	OtherTool string
}

// Plan is the shell script notel audit --fix is about to print.
type Plan struct {
	Mode         Mode
	Suggestions  []Suggestion
	CommandHints []CommandHint
	Conflicts    []Conflict
	Status       Status
}

// ParseShell normalizes a shell name and validates that it is supported.
func ParseShell(name string) (Shell, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "bash":
		return Bash, nil
	case "zsh":
		return Zsh, nil
	case "fish":
		return Fish, nil
	case "powershell", "pwsh", "pwsh.exe", "powershell.exe":
		return PowerShell, nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: bash, fish, powershell, zsh)", name)
	}
}

// DetectShell infers the current shell from the process environment.
func DetectShell() (Shell, error) {
	return Environment{}.DetectShell()
}

// DetectShell infers the current shell from this environment.
//
// SHELL is the Unix signal. On Windows that variable is often empty, so a
// non-empty PSModulePath or a PowerShell ComSpec is the fallback. cmd.exe
// is rejected: its syntax is not one of the shells the fix script can emit.
func (e Environment) DetectShell() (Shell, error) {
	if shellPath := strings.TrimSpace(e.getenv("SHELL")); shellPath != "" {
		return ParseShell(filepath.Base(shellPath))
	}
	if strings.TrimSpace(e.getenv("PSModulePath")) != "" {
		return PowerShell, nil
	}
	if comspec := strings.ToLower(strings.TrimSpace(e.getenv("ComSpec"))); strings.Contains(comspec, "powershell") || strings.Contains(comspec, "pwsh") {
		return PowerShell, nil
	}
	return "", fmt.Errorf("could not detect current shell (supported: bash, fish, powershell, zsh)")
}

// Build returns the env assignments and CLI hints for a fix script.
//
// showAll includes tools that are not installed. mode chooses a minimal
// delta or a full profile. Shared variables such as DO_NOT_TRACK are emitted
// once. The first tool in the result order wins if two tools disagree on the
// value. The disagreement is recorded on Plan.Conflicts for stderr.
func Build(results []check.ToolResult, showAll bool, mode Mode) Plan {
	plan := Plan{Mode: mode}
	seen := make(map[string]Suggestion)

	for _, result := range results {
		if !showAll && !result.Installed {
			continue
		}
		// A delta skips a tool whose opt-out policy already holds.
		// AllPassing also requires the binary, so --all would otherwise
		// treat a missing tool whose variables already match as a failure
		// and append its exports and command hints. Profile mode keeps
		// every visible check so the script can be dropped into a shell rc.
		if mode == ModeDelta && check.PolicySatisfied(result.Tool.OptOut, result.Checks) {
			continue
		}

		for _, envCheck := range checksToSuggest(result, mode) {
			suggestion := Suggestion{
				Category:    result.Tool.Category,
				ToolName:    result.Tool.Name,
				Installed:   result.Installed,
				Current:     envCheck.Value,
				Check:       envCheck.Check,
				TargetValue: preferredValue(envCheck.Check),
			}
			if previous, ok := seen[suggestion.Check.Name]; ok {
				if previous.TargetValue != suggestion.TargetValue {
					plan.Conflicts = append(plan.Conflicts, Conflict{
						Name:      suggestion.Check.Name,
						Kept:      previous.TargetValue,
						KeptTool:  previous.ToolName,
						Other:     suggestion.TargetValue,
						OtherTool: suggestion.ToolName,
					})
				}
				continue
			}
			seen[suggestion.Check.Name] = suggestion
			plan.Suggestions = append(plan.Suggestions, suggestion)
		}

		for _, command := range result.Tool.Commands {
			plan.CommandHints = append(plan.CommandHints, CommandHint{
				Category: result.Tool.Category,
				ToolName: result.Tool.Name,
				Command:  command,
			})
		}
	}

	plan.Status = planStatus(results, showAll, plan)
	return plan
}

// checksToSuggest selects the env results that become export lines.
//
// Profile mode keeps every check. Delta mode on an AllOf tool keeps the
// failing checks. Delta mode on an AnyOf or FirstSet tool keeps the first
// check only. AnyOf failed because none of the alternatives matched, and
// one match is enough. FirstSet failed because the earliest present
// variable is not an opt-out, or none is present; a later alternative
// cannot override a bad earlier value. The registry lists the preferred
// variable first. Callers skip tools whose policy is already satisfied,
// so this slice is only for tools that still need a change.
func checksToSuggest(result check.ToolResult, mode Mode) []check.EnvCheckResult {
	if mode == ModeProfile {
		return result.Checks
	}
	if result.Tool.OptOut == tools.AnyOf || result.Tool.OptOut == tools.FirstSet {
		if len(result.Checks) == 0 {
			return nil
		}
		return result.Checks[:1]
	}
	failing := make([]check.EnvCheckResult, 0, len(result.Checks))
	for _, envCheck := range result.Checks {
		if !envCheck.Passing {
			failing = append(failing, envCheck)
		}
	}
	return failing
}

func planStatus(results []check.ToolResult, showAll bool, plan Plan) Status {
	installed := 0
	for _, result := range results {
		if result.Installed {
			installed++
		}
	}
	// The default audit has nothing to patch when nothing is installed.
	// --all still builds a plan from the catalog, so that case is not this one.
	if !showAll && installed == 0 {
		return StatusNoneInstalled
	}
	if len(plan.Suggestions) == 0 && len(plan.CommandHints) == 0 {
		return StatusAlreadyDisabled
	}
	return StatusFixes
}

// Comment renders a shell-safe comment heading for a category block.
func Comment(category string) string {
	return fmt.Sprintf("# %s Telemetry", category)
}

// Command renders one shell-specific command for the suggested env var.
func Command(shell Shell, suggestion Suggestion) string {
	value := quoteValue(shell, suggestion.TargetValue)
	switch shell {
	case Bash, Zsh:
		return fmt.Sprintf("export %s=%s", suggestion.Check.Name, value)
	case Fish:
		return fmt.Sprintf("set -gx %s %s", suggestion.Check.Name, value)
	case PowerShell:
		return fmt.Sprintf("$Env:%s = %s", suggestion.Check.Name, value)
	default:
		return ""
	}
}

func preferredValue(check tools.EnvCheck) string {
	if len(check.ValidValues) > 0 {
		return check.ValidValues[0]
	}
	return "1"
}

// quoteValue quotes a value for the target shell.
//
// PowerShell single quotes escape an embedded quote by doubling it.
// bash, zsh, and fish single quotes have no escape. A value that contains
// an apostrophe is closed, the apostrophe is added as \', and the quote is
// reopened: 'it'\”s'. The values in the catalog today are 0, 1, true,
// false, off, and ci. The apostrophe form is here so a future value does
// not silently emit a broken script.
func quoteValue(shell Shell, value string) string {
	switch shell {
	case PowerShell:
		return "'" + strings.ReplaceAll(value, "'", "''") + "'"
	default:
		return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
	}
}
