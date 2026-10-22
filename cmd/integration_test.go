package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecuteVersionCommandWritesToStdout(t *testing.T) {
	stdout, stderr, err := executeRootCommand(t, "version")
	if err != nil {
		t.Fatalf("execute version: %v", err)
	}

	for _, want := range []string{
		"notel version",
		"Copyright 2026 Stennie Steneker",
		"License: Apache 2.0",
		"GitHub: https://github.com/stennie/notel",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("version stdout missing %q:\n%s", want, stdout)
		}
	}
	if stderr != "" {
		t.Fatalf("version stderr = %q, want empty", stderr)
	}
}

func TestExecuteHelpCommandWritesStyledHelpToStdout(t *testing.T) {
	stdout, stderr, err := executeRootCommand(t, "help")
	if err != nil {
		t.Fatalf("execute help: %v", err)
	}

	for _, want := range []string{
		"notel  —  DevTools Telemetry Auditor",
		"Available Commands",
		"audit",
		"list",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("help stdout missing %q:\n%s", want, stdout)
		}
	}
	if stderr != "" {
		t.Fatalf("help stderr = %q, want empty", stderr)
	}
}

func TestExecuteAuditFixKeepsTheScriptOnStdout(t *testing.T) {
	t.Cleanup(func() {
		_ = auditCmd.Flags().Set("fix", "")
		_ = auditCmd.Flags().Set("all", "false")
		_ = auditCmd.Flags().Set("verbose", "false")
		_ = auditCmd.Flags().Set("profile", "false")
	})

	stdout, stderr, err := executeRootCommand(t, "audit", "--fix=bash")
	if err != nil {
		t.Fatalf("execute audit --fix: %v", err)
	}
	if strings.Contains(stdout, "Fix Commands") || strings.Contains(stdout, "notel  —") {
		t.Fatalf("audit --fix stdout contained diagnostic output:\n%s", stdout)
	}
	if !strings.Contains(stderr, "Fix Commands") || !strings.Contains(stderr, "Shell: bash") {
		t.Fatalf("audit --fix stderr missing the fix header:\n%s", stderr)
	}
}

func TestExecuteAuditProfileWithoutFixFails(t *testing.T) {
	t.Cleanup(func() {
		_ = auditCmd.Flags().Set("profile", "false")
	})
	_, _, err := executeRootCommand(t, "audit", "--profile")
	if err == nil || !strings.Contains(err.Error(), "--profile") {
		t.Fatalf("execute audit --profile error = %v, want a --profile error", err)
	}
}

func executeRootCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	auditCmd.SetOut(&stdout)
	auditCmd.SetErr(&stderr)
	listCmd.SetOut(&stdout)
	listCmd.SetErr(&stderr)
	versionCmd.SetOut(&stdout)
	versionCmd.SetErr(&stderr)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
		auditCmd.SetOut(nil)
		auditCmd.SetErr(nil)
		listCmd.SetOut(nil)
		listCmd.SetErr(nil)
		versionCmd.SetOut(nil)
		versionCmd.SetErr(nil)
	})

	err := rootCmd.Execute()
	return stdout.String(), stderr.String(), err
}
