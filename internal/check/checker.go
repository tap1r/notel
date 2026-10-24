// Package check performs detection and telemetry-setting audits for developer tools.
package check

import (
	"os"
	"os/exec"
	"strings"

	"github.com/stennie/notel/internal/tools"
)

// Checker reads PATH and the environment. The zero value uses exec.LookPath,
// os.Getenv, and os.LookupEnv. Tests set the field they care about and leave
// the others nil. A nil field falls back to the process, so a test can stub
// PATH lookup and still use t.Setenv for the variables. There is no
// package-level function variable, so tests that do not call t.Setenv can
// run in parallel.
//
// LookupEnv is separate from Getenv. os.Getenv cannot tell an unset variable
// from one set to empty, and a TrimSpace check treats those differently.
// Stubbing Getenv does not change LookupEnv.
type Checker struct {
	LookPath  func(string) (string, error)
	Getenv    func(string) string
	LookupEnv func(string) (string, bool)
}

// NewChecker returns a Checker bound to the current process.
func NewChecker() Checker {
	return Checker{LookPath: exec.LookPath, Getenv: os.Getenv, LookupEnv: os.LookupEnv}
}

func (c Checker) lookPath(file string) (string, error) {
	if c.LookPath == nil {
		return exec.LookPath(file)
	}
	return c.LookPath(file)
}

func (c Checker) getenv(key string) string {
	if c.Getenv == nil {
		return os.Getenv(key)
	}
	return c.Getenv(key)
}

func (c Checker) lookupEnv(key string) (string, bool) {
	if c.LookupEnv == nil {
		return os.LookupEnv(key)
	}
	return c.LookupEnv(key)
}

// Reason explains one environment-variable outcome.
//
// Passing is the boolean the policy combines. Reason tells the formatter
// which sentence to print, because "unset" is a failure for a default-on
// tool and a pass for a default-off tool. Those two situations must not
// share a label.
type Reason int

const (
	// ReasonUnset means the variable is missing and the tool collects
	// telemetry until an opt-out is present.
	ReasonUnset Reason = iota

	// ReasonMatched means the current value is an accepted opt-out.
	ReasonMatched

	// ReasonUnexpected means the variable is set to a value notel does not
	// recognize as an opt-out. The audit reports that as telemetry still
	// enabled. Guessing that an unknown string means "disabled" would hide
	// a mis-typed opt-out.
	ReasonUnexpected

	// ReasonDefaultOff means the variable is unset and Tool.Default is
	// DefaultDisabled. The vendor does not collect until the user opts in,
	// so the missing variable is the private state.
	ReasonDefaultOff
)

// EnvCheckResult holds the outcome of checking a single environment variable.
type EnvCheckResult struct {
	Check   tools.EnvCheck
	Value   string // actual value found in the environment (empty if not set)
	Passing bool
	Reason  Reason
}

// ToolResult holds the full audit result for a single tool.
type ToolResult struct {
	Tool       tools.Tool
	Installed  bool
	BinaryPath string
	Checks     []EnvCheckResult

	// AllPassing is true only when the tool is installed and Tool.OptOut is
	// satisfied. AllOf needs every check to pass. AnyOf needs one check to
	// pass. An opt-out that happens to be present for a tool that is not
	// installed does not count: the summary is about installed tools.
	AllPassing bool
}

// Run checks every tool in the default registry using the process PATH and environment.
func Run() []ToolResult {
	return NewChecker().Run()
}

// Run checks every tool in the default registry.
func (c Checker) Run() []ToolResult {
	return c.RunFor(tools.Registry())
}

// RunFor checks a caller-supplied slice of tools using the process PATH and environment.
func RunFor(toolList []tools.Tool) []ToolResult {
	return NewChecker().RunFor(toolList)
}

// RunFor checks a caller-supplied slice of tools (useful for testing/filtering).
func (c Checker) RunFor(toolList []tools.Tool) []ToolResult {
	results := make([]ToolResult, 0, len(toolList))
	for _, t := range toolList {
		results = append(results, c.auditTool(t))
	}
	return results
}

// Summary holds aggregate counts derived from a slice of ToolResults.
type Summary struct {
	Total     int // total tools in the registry
	Installed int // installed tools
	Passing   int // installed tools whose opt-out policy is satisfied
	Failing   int // installed tools whose opt-out policy is not satisfied
}

// Summarise computes aggregate counts from a result set.
func Summarise(results []ToolResult) Summary {
	s := Summary{Total: len(results)}
	for _, r := range results {
		if r.Installed {
			s.Installed++
			if r.AllPassing {
				s.Passing++
			} else {
				s.Failing++
			}
		}
	}
	return s
}

func (c Checker) auditTool(t tools.Tool) ToolResult {
	result := ToolResult{Tool: t}

	switch {
	case t.AlwaysInstalled:
		// Environment conventions have no executable. They are always in scope.
		result.Installed = true
	case len(t.Detect) == 0:
		// The catalog entry is real, but none of the PATH names we could
		// check would identify this product. Leave it uninstalled so the
		// default audit does not blame a neighboring CLI.
		result.Installed = false
	default:
		for _, name := range t.Detect {
			path, err := c.lookPath(name)
			if err != nil {
				continue
			}
			result.Installed = true
			result.BinaryPath = path
			break
		}
	}

	for _, chk := range t.EnvChecks {
		result.Checks = append(result.Checks, c.checkEnvVar(chk, t.Default))
	}
	result.AllPassing = result.Installed && PolicySatisfied(t.OptOut, result.Checks)
	return result
}

// PolicySatisfied applies Tool.OptOut to the per-variable results.
//
// AllOf is the zero value. Every check has to pass, which matches tools
// whose variables are independent switches.
//
// AnyOf passes on the first success. Use it when the vendor accepts
// alternative opt-outs. A failing alternative is still present in the
// result slice; only the combined AllPassing bit changes.
//
// FirstSet returns the result of the first check whose variable is present.
// Absence is ReasonUnset or ReasonDefaultOff. A present non-opt-out fails
// the tool before later alternatives are considered.
func PolicySatisfied(req tools.Requirement, checks []EnvCheckResult) bool {
	if len(checks) == 0 {
		return true
	}
	switch req {
	case tools.AnyOf:
		for _, cr := range checks {
			if cr.Passing {
				return true
			}
		}
		return false
	case tools.FirstSet:
		for _, cr := range checks {
			if cr.Reason == ReasonUnset || cr.Reason == ReasonDefaultOff {
				continue
			}
			return cr.Passing
		}
		return false
	default:
		for _, cr := range checks {
			if !cr.Passing {
				return false
			}
		}
		return true
	}
}

// checkEnvVar reads one variable and decides whether it satisfies the check.
//
// def is the tool default, not a per-variable default. An unset variable
// passes only when the whole tool is DefaultDisabled. A set-but-unknown
// value fails even then: the user typed something, and it was not an opt-out.
func (c Checker) checkEnvVar(chk tools.EnvCheck, def tools.DefaultState) EnvCheckResult {
	// TrimSpace checks must see a present blank value. os.Getenv returns ""
	// for both "unset" and "set to empty", and gh treats those differently:
	// an absent GH_TELEMETRY falls through to DO_NOT_TRACK, while a blank
	// GH_TELEMETRY is one of the falsey opt-outs. LookupEnv is required
	// because Getenv cannot tell those two states apart.
	var value string
	if chk.TrimSpace {
		raw, present := c.lookupEnv(chk.Name)
		if !present {
			return unsetResult(chk, def)
		}
		value = strings.TrimSpace(raw)
	} else {
		value = c.getenv(chk.Name)
		if value == "" {
			return unsetResult(chk, def)
		}
	}
	if valueAccepted(chk, value) {
		return EnvCheckResult{
			Check:   chk,
			Value:   value,
			Passing: true,
			Reason:  ReasonMatched,
		}
	}
	return EnvCheckResult{
		Check:   chk,
		Value:   value,
		Passing: false,
		Reason:  ReasonUnexpected,
	}
}

func unsetResult(chk tools.EnvCheck, def tools.DefaultState) EnvCheckResult {
	if def == tools.DefaultDisabled {
		return EnvCheckResult{
			Check:   chk,
			Passing: true,
			Reason:  ReasonDefaultOff,
		}
	}
	return EnvCheckResult{
		Check:   chk,
		Passing: false,
		Reason:  ReasonUnset,
	}
}

// valueAccepted reports whether value is an opt-out for chk.
//
// An empty ValidValues list means any non-empty value counts. That rule is
// kept for tools whose vendors say "set the variable" without listing the
// spelling. IgnoreCase uses strings.EqualFold so the registry can list
// "true" once.
func valueAccepted(chk tools.EnvCheck, value string) bool {
	if len(chk.ValidValues) == 0 {
		return true
	}
	for _, valid := range chk.ValidValues {
		if chk.IgnoreCase {
			if strings.EqualFold(value, valid) {
				return true
			}
			continue
		}
		if value == valid {
			return true
		}
	}
	return false
}
