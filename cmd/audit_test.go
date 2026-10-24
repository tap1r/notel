package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stennie/notel/internal/check"
	"github.com/stennie/notel/internal/fix"
	"github.com/stennie/notel/internal/tools"
)

func TestResolveFixShell(t *testing.T) {
	rt := auditRuntime{
		detectShell: func() (fix.Shell, error) { return fix.Zsh, nil },
		parseShell: func(name string) (fix.Shell, error) {
			if name != "powershell" {
				t.Fatalf("parseShell called with %q, want powershell", name)
			}
			return fix.PowerShell, nil
		},
	}

	got, err := resolveFixShell("auto", rt)
	if err != nil {
		t.Fatalf("resolveFixShell(auto) error = %v", err)
	}
	if got != fix.Zsh {
		t.Fatalf("resolveFixShell(auto) = %q, want %q", got, fix.Zsh)
	}

	got, err = resolveFixShell("powershell", rt)
	if err != nil {
		t.Fatalf("resolveFixShell(powershell) error = %v", err)
	}
	if got != fix.PowerShell {
		t.Fatalf("resolveFixShell(powershell) = %q, want %q", got, fix.PowerShell)
	}
}

func TestRunAuditFixUsesDetectedShellShowAllAndDelta(t *testing.T) {
	var capturedShowAll bool
	var capturedMode fix.Mode
	var capturedShell fix.Shell

	rt := auditRuntime{
		run:         func() []check.ToolResult { return []check.ToolResult{{Installed: true}} },
		detectShell: func() (fix.Shell, error) { return fix.Fish, nil },
		parseShell: func(string) (fix.Shell, error) {
			t.Fatal("parseShell should not be called for auto")
			return "", nil
		},
		buildPlan: func(_ []check.ToolResult, showAll bool, mode fix.Mode) fix.Plan {
			capturedShowAll = showAll
			capturedMode = mode
			return fix.Plan{Suggestions: []fix.Suggestion{{ToolName: "Homebrew"}}}
		},
		title:    func(io.Writer) {},
		compact:  func(io.Writer, []check.ToolResult, bool) {},
		detailed: func(io.Writer, []check.ToolResult, bool) {},
		summary:  func(io.Writer, []check.ToolResult) {},
		fixes: func(_ io.Writer, _ io.Writer, shell fix.Shell, plan fix.Plan) {
			capturedShell = shell
			if len(plan.Suggestions) != 1 {
				t.Fatalf("fixes suggestions len = %d, want 1", len(plan.Suggestions))
			}
		},
	}

	if err := runAudit(io.Discard, io.Discard, auditOptions{ShowAll: true, Fix: "auto"}, rt); err != nil {
		t.Fatalf("runAudit() error = %v", err)
	}
	if !capturedShowAll {
		t.Fatal("buildPlan did not receive showAll=true")
	}
	if capturedShell != fix.Fish {
		t.Fatalf("fixes shell = %q, want %q", capturedShell, fix.Fish)
	}
	if capturedMode != fix.ModeDelta {
		t.Fatalf("buildPlan mode = %v, want delta", capturedMode)
	}
}

func TestRunAuditExplicitShellBypassesDetection(t *testing.T) {
	called := false
	rt := auditRuntime{
		run: func() []check.ToolResult { return nil },
		detectShell: func() (fix.Shell, error) {
			t.Fatal("detectShell should not be called for an explicit shell")
			return "", nil
		},
		parseShell: func(name string) (fix.Shell, error) {
			if name != "bash" {
				t.Fatalf("parseShell called with %q, want bash", name)
			}
			return fix.Bash, nil
		},
		buildPlan: func([]check.ToolResult, bool, fix.Mode) fix.Plan { return fix.Plan{} },
		title:     func(io.Writer) {},
		fixes: func(_ io.Writer, _ io.Writer, shell fix.Shell, _ fix.Plan) {
			called = true
			if shell != fix.Bash {
				t.Fatalf("fixes shell = %q, want %q", shell, fix.Bash)
			}
		},
	}

	if err := runAudit(io.Discard, io.Discard, auditOptions{Fix: "bash"}, rt); err != nil {
		t.Fatalf("runAudit() error = %v", err)
	}
	if !called {
		t.Fatal("fixes was not called")
	}
}

func TestRunAuditStdoutContainsOnlyCommands(t *testing.T) {
	rt := auditRuntime{
		run:        func() []check.ToolResult { return nil },
		parseShell: func(string) (fix.Shell, error) { return fix.Bash, nil },
		buildPlan: func([]check.ToolResult, bool, fix.Mode) fix.Plan {
			return fix.Plan{Status: fix.StatusFixes, Suggestions: []fix.Suggestion{
				{Category: "Package Manager", Check: tools.EnvCheck{Name: "TEST_VAR"}, TargetValue: "1"},
				{Category: "Package Manager", Check: tools.EnvCheck{Name: "ALREADY_SET"}, Current: "1", TargetValue: "1"},
				{Category: "Framework", Check: tools.EnvCheck{Name: "FRAMEWORK_VAR"}, TargetValue: "0"},
			}}
		},
		title:   func(io.Writer) {},
		summary: func(io.Writer, []check.ToolResult) {},
		fixes:   outputPrintFix,
	}

	var stdout, stderr bytes.Buffer
	if err := runAudit(&stdout, &stderr, auditOptions{Fix: "bash"}, rt); err != nil {
		t.Fatalf("runAudit() error = %v", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "# Package Manager Telemetry\nexport TEST_VAR='1'\nexport ALREADY_SET='1'\n\n# Framework Telemetry\nexport FRAMEWORK_VAR='0'" {
		t.Fatalf("stdout = %q, want the shell script only", got)
	}
	if !strings.Contains(stderr.String(), "Fix Commands") {
		t.Fatalf("stderr missing contextual output:\n%s", stderr.String())
	}
	if strings.Contains(stdout.String(), "Fix Commands") {
		t.Fatalf("stdout contained contextual output:\n%s", stdout.String())
	}
}

func TestRunAuditProfileRequiresFix(t *testing.T) {
	rt := auditRuntime{run: func() []check.ToolResult { return nil }}
	err := runAudit(io.Discard, io.Discard, auditOptions{Profile: true}, rt)
	if err == nil || !strings.Contains(err.Error(), "--profile") {
		t.Fatalf("runAudit() error = %v, want a --profile error", err)
	}
}

func TestRunAuditChoosesCompactOrDetailedView(t *testing.T) {
	rt := auditRuntime{
		run:     func() []check.ToolResult { return []check.ToolResult{{Installed: true}} },
		title:   func(io.Writer) {},
		summary: func(io.Writer, []check.ToolResult) {},
	}

	tests := []struct {
		name        string
		opts        auditOptions
		wantCompact int
		wantDetail  int
	}{
		{name: "all without verbose", opts: auditOptions{ShowAll: true}, wantCompact: 1},
		{name: "all with verbose", opts: auditOptions{ShowAll: true, Verbose: true}, wantDetail: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compact, detail := 0, 0
			rt.compact = func(_ io.Writer, _ []check.ToolResult, showAll bool) {
				compact++
				if showAll != tt.opts.ShowAll {
					t.Fatalf("compact showAll = %v", showAll)
				}
			}
			rt.detailed = func(_ io.Writer, _ []check.ToolResult, showAll bool) {
				detail++
				if showAll != tt.opts.ShowAll {
					t.Fatalf("detailed showAll = %v", showAll)
				}
			}
			if err := runAudit(io.Discard, io.Discard, tt.opts, rt); err != nil {
				t.Fatalf("runAudit() error = %v", err)
			}
			if compact != tt.wantCompact || detail != tt.wantDetail {
				t.Fatalf("compact=%d detail=%d, want %d and %d", compact, detail, tt.wantCompact, tt.wantDetail)
			}
		})
	}
}

func TestAuditFlagsReadIntoOptions(t *testing.T) {
	t.Cleanup(func() {
		_ = auditCmd.Flags().Set("all", "false")
		_ = auditCmd.Flags().Set("verbose", "false")
		_ = auditCmd.Flags().Set("fix", "")
		_ = auditCmd.Flags().Set("profile", "false")
	})
	if err := auditCmd.Flags().Set("all", "true"); err != nil {
		t.Fatal(err)
	}
	if err := auditCmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatal(err)
	}
	if err := auditCmd.Flags().Set("fix", "zsh"); err != nil {
		t.Fatal(err)
	}
	if err := auditCmd.Flags().Set("profile", "true"); err != nil {
		t.Fatal(err)
	}

	opts, err := auditOptionsFrom(auditCmd)
	if err != nil {
		t.Fatalf("auditOptionsFrom() error = %v", err)
	}
	if !opts.ShowAll || !opts.Verbose || !opts.Profile || opts.Fix != "zsh" {
		t.Fatalf("auditOptionsFrom() = %#v", opts)
	}
}

func TestAuditFixFlagHasFShorthand(t *testing.T) {
	flag := auditCmd.Flags().Lookup("fix")
	if flag == nil {
		t.Fatal("fix flag is not registered")
	}
	if flag.Shorthand != "f" {
		t.Fatalf("fix flag shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

// outputPrintFix is the real printer, named so the runtime literal stays readable.
var outputPrintFix = productionAuditRuntime().fixes
