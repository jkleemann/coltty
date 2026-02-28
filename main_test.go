package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func executeCommand(args ...string) (string, string, error) {
	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	rootCmd.SetArgs(args)
	err := rootCmd.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var outBuf, errBuf bytes.Buffer
	outBuf.ReadFrom(rOut)
	errBuf.ReadFrom(rErr)

	return outBuf.String(), errBuf.String(), err
}

func TestInitZsh(t *testing.T) {
	stdout, _, err := executeCommand("init", "zsh")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "coltty_chpwd") {
		t.Error("expected zsh hook output")
	}
}

func TestInitBash(t *testing.T) {
	stdout, _, err := executeCommand("init", "bash")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "PROMPT_COMMAND") {
		t.Error("expected bash hook output")
	}
}

func TestApplyDryRun(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".coltty.toml"), []byte(`scheme = "test"`), 0644)

	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	stdout, _, err := executeCommand("apply", "--dry-run")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "Source:") {
		t.Error("expected scheme output in dry-run mode")
	}
}

func TestShowCommand(t *testing.T) {
	dir := t.TempDir()

	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	stdout, _, err := executeCommand("show")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "Source:") {
		t.Error("expected scheme output from show command")
	}
	if !strings.Contains(stdout, "(default)") {
		t.Error("expected default source when no .coltty.toml exists")
	}
}

func TestSchemesCommandNoConfig(t *testing.T) {
	globalConfigPathOverride = filepath.Join(t.TempDir(), "nonexistent", "config.toml")
	defer func() { globalConfigPathOverride = "" }()

	stdout, _, err := executeCommand("schemes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "No schemes defined") {
		t.Error("expected 'No schemes defined' message")
	}
}

func TestSchemesCommandWithConfig(t *testing.T) {
	configDir := t.TempDir()

	config := `
[default]
scheme = "calm"

[schemes.calm]
foreground = "#c0caf5"
background = "#1a1b26"
cursor = "#c0caf5"

[schemes.danger]
foreground = "#f8f8f2"
background = "#3b0a0a"
cursor = "#ff5555"
`
	configPath := filepath.Join(configDir, "config.toml")
	os.WriteFile(configPath, []byte(config), 0644)

	globalConfigPathOverride = configPath
	defer func() { globalConfigPathOverride = "" }()

	stdout, _, err := executeCommand("schemes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "calm") {
		t.Error("expected 'calm' scheme in output")
	}
	if !strings.Contains(stdout, "danger") {
		t.Error("expected 'danger' scheme in output")
	}
	if !strings.Contains(stdout, "(default)") {
		t.Error("expected default marker on calm scheme")
	}
}
