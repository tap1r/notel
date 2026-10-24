package fix

import (
	"testing"

	"github.com/stennie/notel/internal/check"
	"github.com/stennie/notel/internal/tools"
)

func TestParseShell(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Shell
		wantErr bool
	}{
		{name: "bash", input: "bash", want: Bash},
		{name: "zsh", input: "zsh", want: Zsh},
		{name: "fish", input: "fish", want: Fish},
		{name: "pwsh", input: "pwsh", want: PowerShell},
		{name: "powershell exe", input: "powershell.exe", want: PowerShell},
		{name: "invalid", input: "cmd", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseShell(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseShell() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseShell() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseShell() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectShell(t *testing.T) {
	originalGetEnv := getEnv
	t.Cleanup(func() { getEnv = originalGetEnv })

	tests := []struct {
		name    string
		env     map[string]string
		want    Shell
		wantErr bool
	}{
		{name: "shell path zsh", env: map[string]string{"SHELL": "/bin/zsh"}, want: Zsh},
		{name: "powershell via module path", env: map[string]string{"PSModulePath": "x"}, want: PowerShell},
		{name: "undetected", env: map[string]string{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getEnv = func(key string) string { return tt.env[key] }
			got, err := DetectShell()
			if tt.wantErr {
				if err == nil {
					t.Fatal("DetectShell() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("DetectShell() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("DetectShell() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildDeltaSkipsPassingToolsAndDedupes(t *testing.T) {
	results := []check.ToolResult{
		{
			Tool:       tools.Tool{Name: "Do Not Track", Category: "All The Things"},
			Installed:  true,
			AllPassing: false,
			Checks: []check.EnvCheckResult{
				{Check: tools.EnvCheck{Name: "DO_NOT_TRACK", ValidValues: []string{"1"}}, Passing: false},
			},
		},
		{
			Tool:       tools.Tool{Name: "Homebrew", Category: "Package Manager", Commands: []string{"brew analytics off"}},
			Installed:  true,
			AllPassing: true,
			Checks: []check.EnvCheckResult{
				{Check: tools.EnvCheck{Name: "HOMEBREW_NO_ANALYTICS", ValidValues: []string{"1"}}, Value: "1", Passing: true},
			},
		},
		{
			Tool:       tools.Tool{Name: "GitHub CLI", Category: "Cloud", OptOut: tools.AnyOf, Commands: []string{"gh config set telemetry disabled"}},
			Installed:  true,
			AllPassing: false,
			Checks: []check.EnvCheckResult{
				{Check: tools.EnvCheck{Name: "GH_TELEMETRY", ValidValues: []string{"false"}}, Passing: false},
				{Check: tools.EnvCheck{Name: "DO_NOT_TRACK", ValidValues: []string{"1"}}, Passing: false},
			},
		},
		{
			Tool:      tools.Tool{Name: "Missing"},
			Installed: false,
			Checks: []check.EnvCheckResult{
				{Check: tools.EnvCheck{Name: "MISSING", ValidValues: []string{"0"}}, Passing: false},
			},
		},
	}

	got := Build(results, false, ModeDelta)
	if got.Status != StatusFixes {
		t.Fatalf("delta status = %v, want fixes", got.Status)
	}
	if len(got.Suggestions) != 2 {
		t.Fatalf("delta suggestions = %#v, want DO_NOT_TRACK once and GH_TELEMETRY", got.Suggestions)
	}
	if got.Suggestions[0].Check.Name != "DO_NOT_TRACK" || got.Suggestions[1].Check.Name != "GH_TELEMETRY" {
		t.Fatalf("delta suggestion names = %q, %q", got.Suggestions[0].Check.Name, got.Suggestions[1].Check.Name)
	}
	if len(got.CommandHints) != 1 || got.CommandHints[0].Command != "gh config set telemetry disabled" {
		t.Fatalf("delta command hints = %#v, want only the failing tool", got.CommandHints)
	}

	// Passing tools stay in a profile, and the shared variable is still once.
	profile := Build(results, false, ModeProfile)
	if len(profile.Suggestions) != 3 {
		t.Fatalf("profile suggestions len = %d, want DO_NOT_TRACK, HOMEBREW_NO_ANALYTICS, GH_TELEMETRY", len(profile.Suggestions))
	}
	if profile.Suggestions[1].Check.Name != "HOMEBREW_NO_ANALYTICS" {
		t.Fatalf("profile second suggestion = %q, want the passing Homebrew variable", profile.Suggestions[1].Check.Name)
	}
}

func TestBuildReportsValueConflictsAndEmptyReasons(t *testing.T) {
	conflict := Build([]check.ToolResult{
		{
			Tool:       tools.Tool{Name: "First"},
			Installed:  true,
			AllPassing: false,
			Checks: []check.EnvCheckResult{
				{Check: tools.EnvCheck{Name: "SHARED", ValidValues: []string{"1"}}, Passing: false},
			},
		},
		{
			Tool:       tools.Tool{Name: "Second"},
			Installed:  true,
			AllPassing: false,
			Checks: []check.EnvCheckResult{
				{Check: tools.EnvCheck{Name: "SHARED", ValidValues: []string{"true"}}, Passing: false},
			},
		},
	}, false, ModeDelta)
	if len(conflict.Suggestions) != 1 || conflict.Suggestions[0].TargetValue != "1" {
		t.Fatalf("conflict suggestions = %#v, want the first value kept", conflict.Suggestions)
	}
	if len(conflict.Conflicts) != 1 || conflict.Conflicts[0].Other != "true" || conflict.Conflicts[0].OtherTool != "Second" {
		t.Fatalf("conflicts = %#v", conflict.Conflicts)
	}

	none := Build(nil, false, ModeDelta)
	if none.Status != StatusNoneInstalled {
		t.Fatalf("empty audit status = %v, want none installed", none.Status)
	}

	passing := Build([]check.ToolResult{{
		Tool:       tools.Tool{Name: "Homebrew"},
		Installed:  true,
		AllPassing: true,
		Checks: []check.EnvCheckResult{
			{Check: tools.EnvCheck{Name: "HOMEBREW_NO_ANALYTICS", ValidValues: []string{"1"}}, Value: "1", Passing: true},
		},
	}}, false, ModeDelta)
	if passing.Status != StatusAlreadyDisabled {
		t.Fatalf("passing audit status = %v, want already disabled", passing.Status)
	}
}

func TestBuildDeltaSkipsSatisfiedPolicyEvenWhenNotInstalled(t *testing.T) {
	homebrew := check.ToolResult{
		Tool: tools.Tool{
			Name:     "Homebrew",
			Commands: []string{"brew analytics off"},
		},
		Installed:  false,
		AllPassing: false,
		Checks: []check.EnvCheckResult{
			{Check: tools.EnvCheck{Name: "HOMEBREW_NO_ANALYTICS", ValidValues: []string{"1"}}, Value: "1", Passing: true, Reason: check.ReasonMatched},
		},
	}
	github := check.ToolResult{
		Tool: tools.Tool{
			Name:     "GitHub CLI",
			OptOut:   tools.FirstSet,
			Commands: []string{"gh config set telemetry disabled"},
		},
		Installed:  false,
		AllPassing: false,
		Checks: []check.EnvCheckResult{
			{Check: tools.EnvCheck{Name: "GH_TELEMETRY", ValidValues: []string{"false"}}, Passing: false, Reason: check.ReasonUnset},
			{Check: tools.EnvCheck{Name: "DO_NOT_TRACK", ValidValues: []string{"1"}}, Value: "1", Passing: true, Reason: check.ReasonMatched},
		},
	}
	gcloud := check.ToolResult{
		Tool: tools.Tool{
			Name:     "Google Cloud SDK",
			Default:  tools.DefaultDisabled,
			Commands: []string{"gcloud config set disable_usage_reporting true"},
		},
		Installed: false,
		Checks: []check.EnvCheckResult{
			{Check: tools.EnvCheck{Name: "CLOUDSDK_CORE_DISABLE_USAGE_REPORTING", ValidValues: []string{"true"}}, Passing: true, Reason: check.ReasonDefaultOff},
		},
	}

	got := Build([]check.ToolResult{homebrew, github, gcloud}, true, ModeDelta)
	if len(got.Suggestions) != 0 || len(got.CommandHints) != 0 {
		t.Fatalf("satisfied --all delta = suggestions %#v hints %#v, want neither", got.Suggestions, got.CommandHints)
	}
	if got.Status != StatusAlreadyDisabled {
		t.Fatalf("status = %v, want already disabled", got.Status)
	}

	profile := Build([]check.ToolResult{homebrew}, true, ModeProfile)
	if len(profile.Suggestions) != 1 || len(profile.CommandHints) != 1 {
		t.Fatalf("profile = suggestions %d hints %d, want the variable and the brew hint", len(profile.Suggestions), len(profile.CommandHints))
	}

	// A set non-opt-out still fails, and a later match cannot override it.
	failing := github
	failing.Checks = []check.EnvCheckResult{
		{Check: tools.EnvCheck{Name: "GH_TELEMETRY", ValidValues: []string{"false"}}, Value: "1", Passing: false, Reason: check.ReasonUnexpected},
		{Check: tools.EnvCheck{Name: "DO_NOT_TRACK", ValidValues: []string{"1"}}, Value: "1", Passing: true, Reason: check.ReasonMatched},
	}
	delta := Build([]check.ToolResult{failing}, true, ModeDelta)
	if len(delta.Suggestions) != 1 || delta.Suggestions[0].Check.Name != "GH_TELEMETRY" || delta.Suggestions[0].TargetValue != "false" {
		t.Fatalf("FirstSet delta = %#v, want only GH_TELEMETRY=false", delta.Suggestions)
	}
	if len(delta.CommandHints) != 1 || delta.CommandHints[0].Command != "gh config set telemetry disabled" {
		t.Fatalf("FirstSet hints = %#v, want the gh config hint", delta.CommandHints)
	}
}

func TestQuoteValueEscapesApostrophes(t *testing.T) {
	suggestion := Suggestion{
		Check:       tools.EnvCheck{Name: "EXAMPLE"},
		TargetValue: "it's",
	}
	if got := Command(Bash, suggestion); got != `export EXAMPLE='it'\''s'` {
		t.Fatalf("bash quote = %q", got)
	}
	if got := Command(Fish, suggestion); got != `set -gx EXAMPLE 'it'\''s'` {
		t.Fatalf("fish quote = %q", got)
	}
	if got := Command(PowerShell, suggestion); got != `$Env:EXAMPLE = 'it''s'` {
		t.Fatalf("powershell quote = %q", got)
	}
}

func TestCommand(t *testing.T) {
	suggestion := Suggestion{
		Check:       tools.EnvCheck{Name: "HOMEBREW_NO_ANALYTICS"},
		TargetValue: "1",
	}

	tests := []struct {
		shell Shell
		want  string
	}{
		{shell: Bash, want: "export HOMEBREW_NO_ANALYTICS='1'"},
		{shell: Zsh, want: "export HOMEBREW_NO_ANALYTICS='1'"},
		{shell: Fish, want: "set -gx HOMEBREW_NO_ANALYTICS '1'"},
		{shell: PowerShell, want: "$Env:HOMEBREW_NO_ANALYTICS = '1'"},
	}

	for _, tt := range tests {
		t.Run(string(tt.shell), func(t *testing.T) {
			if got := Command(tt.shell, suggestion); got != tt.want {
				t.Fatalf("Command() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComment(t *testing.T) {
	if got := Comment("Package Manager"); got != "# Package Manager Telemetry" {
		t.Fatalf("Comment() = %q, want %q", got, "# Package Manager Telemetry")
	}
}
