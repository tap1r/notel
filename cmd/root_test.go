package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommandSilencesErrorsAndUsage(t *testing.T) {
	if !rootCmd.SilenceErrors {
		t.Fatal("rootCmd.SilenceErrors = false, want true")
	}
	if !rootCmd.SilenceUsage {
		t.Fatal("rootCmd.SilenceUsage = false, want true")
	}
}

func TestLeafCommandsRejectUnexpectedArgs(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "audit", cmd: auditCmd},
		{name: "list", cmd: listCmd},
		{name: "version", cmd: versionCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cmd.Args(tt.cmd, []string{"unexpected"}); err == nil {
				t.Fatal("Args accepted unexpected positional argument")
			}
		})
	}
}

func TestListCommandHasLsAlias(t *testing.T) {
	found := false
	for _, alias := range listCmd.Aliases {
		if alias == "ls" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("listCmd.Aliases = %#v, want to include %q", listCmd.Aliases, "ls")
	}
}

func TestRootHelpUsesStyledHeader(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	if err := rootCmd.Help(); err != nil {
		t.Fatalf("rootCmd.Help() error = %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		"notel  —  DevTools Telemetry Auditor",
		"Usage",
		"Available Commands",
		"audit",
		"Copyright 2026 Stennie Steneker",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("styled help output missing %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "Author: Stennie Steneker") {
		t.Fatalf("styled help output still contains author line:\n%s", output)
	}
}

func TestVersionCommandIncludesMetadata(t *testing.T) {
	var buf bytes.Buffer
	versionCmd.SetOut(&buf)
	versionCmd.SetErr(&buf)

	versionCmd.Run(versionCmd, nil)

	output := buf.String()
	for _, want := range []string{
		"notel version",
		"Copyright 2026 Stennie Steneker",
		"License: Apache 2.0",
		"GitHub: https://github.com/stennie/notel",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("version output missing %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "Author: Stennie Steneker") {
		t.Fatalf("version output still contains author line:\n%s", output)
	}
}
