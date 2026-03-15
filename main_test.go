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
	// With no user config, built-in schemes should still be listed.
	if !strings.Contains(stdout, "gruvbox") {
		t.Error("expected 'gruvbox' built-in scheme in output")
	}
	if !strings.Contains(stdout, "(built-in)") {
		t.Error("expected '(built-in)' marker in output")
	}
}

func TestSetCommandCreatesConfig(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	globalConfigPathOverride = filepath.Join(t.TempDir(), "nonexistent", "config.toml")
	defer func() { globalConfigPathOverride = "" }()

	_, stderr, err := executeCommand("set", "dracula")
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".coltty.toml"))
	if err != nil {
		t.Fatal("expected .coltty.toml to be created:", err)
	}
	content := string(data)
	if !strings.Contains(content, `scheme = "dracula"`) {
		t.Errorf("expected scheme = \"dracula\", got:\n%s", content)
	}
	if !strings.Contains(stderr, `set scheme "dracula"`) {
		t.Errorf("expected confirmation on stderr, got: %s", stderr)
	}
}

func TestSetCommandWithoutArgsStartsPicker(t *testing.T) {
	called := false
	old := interactiveSetRunner
	interactiveSetRunner = func() error {
		called = true
		return nil
	}
	defer func() { interactiveSetRunner = old }()

	_, _, err := executeCommand("set")
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected interactive picker to start")
	}
}

func TestSetCommandInline(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	globalConfigPathOverride = filepath.Join(t.TempDir(), "nonexistent", "config.toml")
	defer func() { globalConfigPathOverride = "" }()
	defer func() { setInline = false }()

	_, _, err := executeCommand("set", "dracula", "--inline")
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".coltty.toml"))
	if err != nil {
		t.Fatal("expected .coltty.toml to be created:", err)
	}
	content := string(data)
	if strings.Contains(content, `scheme = "dracula"`) {
		t.Error("inline mode should not write a scheme reference")
	}
	if !strings.Contains(content, "[overrides]") {
		t.Error("expected [overrides] section in inline mode")
	}
	if !strings.Contains(content, `foreground = "#f8f8f2"`) {
		t.Error("expected dracula foreground color")
	}
	if !strings.Contains(content, `background = "#282a36"`) {
		t.Error("expected dracula background color")
	}
	if !strings.Contains(content, `"#ff5555"`) {
		t.Error("expected palette colors")
	}
}

func TestSetCommandInlineStillWritesOverrides(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	globalConfigPathOverride = filepath.Join(t.TempDir(), "nonexistent", "config.toml")
	defer func() { globalConfigPathOverride = "" }()
	defer func() { setInline = false }()

	_, _, err := executeCommand("set", "dracula", "--inline")
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".coltty.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "[overrides]") {
		t.Fatal("expected inline overrides")
	}
}

func TestSetCommandRejectsUnknown(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	globalConfigPathOverride = filepath.Join(t.TempDir(), "nonexistent", "config.toml")
	defer func() { globalConfigPathOverride = "" }()

	_, _, err := executeCommand("set", "nonexistent-scheme")
	if err == nil {
		t.Fatal("expected error for unknown scheme")
	}

	// File should not be created.
	if _, statErr := os.Stat(filepath.Join(dir, ".coltty.toml")); statErr == nil {
		t.Error("expected no .coltty.toml for unknown scheme")
	}
}

func TestSetCommandOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	globalConfigPathOverride = filepath.Join(t.TempDir(), "nonexistent", "config.toml")
	defer func() { globalConfigPathOverride = "" }()

	// Create an existing file.
	os.WriteFile(filepath.Join(dir, ".coltty.toml"), []byte(`scheme = "nord"`), 0644)

	_, stderr, err := executeCommand("set", "dracula")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stderr, "overwriting") {
		t.Error("expected overwrite warning on stderr")
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".coltty.toml"))
	if !strings.Contains(string(data), `scheme = "dracula"`) {
		t.Error("expected file to be overwritten with new scheme")
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

[schemes.dracula]
foreground = "#custom"
background = "#override"
cursor = "#user"
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
	if !strings.Contains(stdout, "gruvbox") {
		t.Error("expected 'gruvbox' built-in scheme in output")
	}
	if !strings.Contains(stdout, "(default)") {
		t.Error("expected default marker on calm scheme")
	}
	// dracula is both built-in and user-defined, so it should show (override)
	if !strings.Contains(stdout, "(override)") {
		t.Error("expected '(override)' marker for user-overridden built-in scheme")
	}
}
