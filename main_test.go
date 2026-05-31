package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestRootCmd() *cobra.Command {
	return newRootCmd()
}

func executeCommand(cmd *cobra.Command, args ...string) (string, string, error) {
	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	cmd.SetArgs(args)
	err := cmd.Execute()

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
	stdout, _, err := executeCommand(newTestRootCmd(), "init", "zsh")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "coltty_chpwd") {
		t.Error("expected zsh hook output")
	}
}

func TestInitBash(t *testing.T) {
	stdout, _, err := executeCommand(newTestRootCmd(), "init", "bash")
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

	stdout, _, err := executeCommand(newTestRootCmd(), "apply", "--dry-run")
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

	stdout, _, err := executeCommand(newTestRootCmd(), "show")
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

	stdout, _, err := executeCommand(newTestRootCmd(), "schemes")
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

	_, stderr, err := executeCommand(newTestRootCmd(), "set", "dracula")
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

func TestSetCommandInline(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	globalConfigPathOverride = filepath.Join(t.TempDir(), "nonexistent", "config.toml")
	defer func() { globalConfigPathOverride = "" }()

	_, _, err := executeCommand(newTestRootCmd(), "set", "dracula", "--inline")
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

func TestSetCommandRejectsUnknown(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	globalConfigPathOverride = filepath.Join(t.TempDir(), "nonexistent", "config.toml")
	defer func() { globalConfigPathOverride = "" }()

	_, _, err := executeCommand(newTestRootCmd(), "set", "nonexistent-scheme")
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

	_, stderr, err := executeCommand(newTestRootCmd(), "set", "dracula")
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

	stdout, _, err := executeCommand(newTestRootCmd(), "schemes")
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

func TestLookupSchemeNotFound(t *testing.T) {
	_, ok := lookupScheme("nonexistent", nil)
	if ok {
		t.Error("expected lookupScheme to return false for unknown scheme")
	}

	globalCfg := &GlobalConfig{
		Schemes: map[string]Scheme{
			"custom": {Foreground: "#111", Background: "#222", Cursor: "#333"},
		},
	}
	_, ok = lookupScheme("also-nonexistent", globalCfg)
	if ok {
		t.Error("expected lookupScheme to return false for unknown scheme even with global config")
	}
}

func TestFormatInlineConfigEmptyPalette(t *testing.T) {
	scheme := Scheme{
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Cursor:     "#c0caf5",
		Palette:    []string{},
	}
	result := formatInlineConfig("test", scheme)
	if strings.Contains(result, "palette = [") {
		t.Error("expected no palette section for empty palette")
	}
	if !strings.Contains(result, `foreground = "#c0caf5"`) {
		t.Error("expected foreground in output")
	}
}

func TestToAdapterSchemeAllExtras(t *testing.T) {
	resolved := &ResolvedScheme{
		Foreground:          "#c0caf5",
		Background:          "#1a1b26",
		Cursor:              "#c0caf5",
		Palette:             []string{"#111"},
		SchemeName:          "test",
		Bold:                "#ffffff",
		SelectionForeground: "#000000",
		SelectionBackground: "#111111",
		Tab:                 "#222222",
		ItermPreset:         "Preset Name",
		TerminalAppProfile:  "Profile Name",
	}
	adapterScheme := toAdapterScheme(resolved)
	if adapterScheme.Foreground != "#c0caf5" {
		t.Errorf("expected foreground '#c0caf5', got %q", adapterScheme.Foreground)
	}
	if adapterScheme.Name != "test" {
		t.Errorf("expected name 'test', got %q", adapterScheme.Name)
	}
	if adapterScheme.Extras == nil {
		t.Fatal("expected extras to be set")
	}
	if adapterScheme.Extras["bold"] != "#ffffff" {
		t.Errorf("expected bold extra, got %q", adapterScheme.Extras["bold"])
	}
	if adapterScheme.Extras["selection_foreground"] != "#000000" {
		t.Errorf("expected selection_foreground extra, got %q", adapterScheme.Extras["selection_foreground"])
	}
	if adapterScheme.Extras["selection_background"] != "#111111" {
		t.Errorf("expected selection_background extra, got %q", adapterScheme.Extras["selection_background"])
	}
	if adapterScheme.Extras["tab"] != "#222222" {
		t.Errorf("expected tab extra, got %q", adapterScheme.Extras["tab"])
	}
	if adapterScheme.Extras["iterm_preset"] != "Preset Name" {
		t.Errorf("expected iterm_preset extra, got %q", adapterScheme.Extras["iterm_preset"])
	}
	if adapterScheme.Extras["terminal_app_profile"] != "Profile Name" {
		t.Errorf("expected terminal_app_profile extra, got %q", adapterScheme.Extras["terminal_app_profile"])
	}
}

func TestToAdapterSchemeNoExtras(t *testing.T) {
	resolved := &ResolvedScheme{
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Cursor:     "#c0caf5",
		Palette:    []string{"#111"},
		SchemeName: "minimal",
	}
	adapterScheme := toAdapterScheme(resolved)
	if adapterScheme.Extras != nil {
		t.Errorf("expected nil extras when no extended colors set, got %v", adapterScheme.Extras)
	}
}

func TestPrintSchemeWithPalette(t *testing.T) {
	s := &ResolvedScheme{
		Source:     "/test/path",
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Cursor:     "#c0caf5",
		Palette:    []string{"#111", "#222", "#333"},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printScheme(s)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	expected := "Source:     /test/path\nForeground: #c0caf5\nBackground: #1a1b26\nCursor:     #c0caf5\nPalette:    #111, #222, #333\n"
	if out != expected {
		t.Errorf("unexpected output:\ngot:\n%q\nwant:\n%q", out, expected)
	}
}

func TestPrintSchemeNoPalette(t *testing.T) {
	s := &ResolvedScheme{
		Source:     "(default)",
		Foreground: "#ebdbb2",
		Background: "#282828",
		Cursor:     "#ebdbb2",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printScheme(s)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	expected := "Source:     (default)\nForeground: #ebdbb2\nBackground: #282828\nCursor:     #ebdbb2\n"
	if out != expected {
		t.Errorf("unexpected output:\ngot:\n%q\nwant:\n%q", out, expected)
	}
}

func TestToAdapterSchemePartialExtras(t *testing.T) {
	resolved := &ResolvedScheme{
		Foreground:         "#c0caf5",
		Background:         "#1a1b26",
		Cursor:             "#c0caf5",
		SchemeName:         "partial",
		Bold:               "#ffffff",
		TerminalAppProfile: "Basic",
	}
	adapterScheme := toAdapterScheme(resolved)
	if adapterScheme.Extras == nil {
		t.Fatal("expected extras to be set")
	}
	if len(adapterScheme.Extras) != 2 {
		t.Errorf("expected 2 extras, got %d", len(adapterScheme.Extras))
	}
	if adapterScheme.Extras["bold"] != "#ffffff" {
		t.Errorf("expected bold extra, got %q", adapterScheme.Extras["bold"])
	}
	if adapterScheme.Extras["terminal_app_profile"] != "Basic" {
		t.Errorf("expected terminal_app_profile extra, got %q", adapterScheme.Extras["terminal_app_profile"])
	}
	if _, ok := adapterScheme.Extras["selection_foreground"]; ok {
		t.Error("expected selection_foreground to not be set")
	}
}

func TestFormatInlineConfigWithPalette(t *testing.T) {
	scheme := Scheme{
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Cursor:     "#c0caf5",
		Palette:    []string{"#111", "#222", "#333", "#444", "#555"},
	}
	result := formatInlineConfig("test", scheme)

	expected := `# Generated from scheme "test"

[overrides]
foreground = "#c0caf5"
background = "#1a1b26"
cursor = "#c0caf5"
palette = [
    "#111", "#222", "#333", "#444",
    "#555"
]
`
	if result != expected {
		t.Errorf("unexpected output:\ngot:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatInlineConfigWithPaletteExactMultiple(t *testing.T) {
	scheme := Scheme{
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Cursor:     "#c0caf5",
		Palette:    []string{"#111", "#222", "#333", "#444"},
	}
	result := formatInlineConfig("test", scheme)

	expected := `# Generated from scheme "test"

[overrides]
foreground = "#c0caf5"
background = "#1a1b26"
cursor = "#c0caf5"
palette = [
    "#111", "#222", "#333", "#444"
]
`
	if result != expected {
		t.Errorf("unexpected output:\ngot:\n%q\nwant:\n%q", result, expected)
	}
}
