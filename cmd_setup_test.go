package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupTerminalAppCommand(t *testing.T) {
	// No user config — only built-in schemes.
	globalConfigPathOverride = filepath.Join(t.TempDir(), "nonexistent", "config.toml")
	defer func() { globalConfigPathOverride = "" }()

	var scripts []string
	setupRunAppleScript = func(script string) error {
		scripts = append(scripts, script)
		return nil
	}
	defer func() { setupRunAppleScript = nil }()

	_, stderr, err := executeCommand("setup", "terminal-app")
	if err != nil {
		t.Fatal(err)
	}

	builtins := BuiltinSchemes()
	if len(scripts) != len(builtins) {
		t.Errorf("expected %d scripts (one per built-in scheme), got %d", len(builtins), len(scripts))
	}

	// Every script should create a profile but NOT switch to it.
	for _, script := range scripts {
		if !strings.Contains(script, "is not in profileNames") {
			t.Errorf("expected profile creation check in script:\n%s", script)
		}
		if strings.Contains(script, "set current settings of front window") {
			t.Errorf("setup script should NOT switch profile:\n%s", script)
		}
	}

	// Check summary output.
	if !strings.Contains(stderr, "created/updated") {
		t.Errorf("expected summary on stderr, got: %s", stderr)
	}

	// Check that each built-in scheme name appears in the checkmark output.
	for name := range builtins {
		if !strings.Contains(stderr, name) {
			t.Errorf("expected scheme %q in stderr output", name)
		}
	}
}

func TestSetupTerminalAppWithUserSchemes(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.toml")

	config := `
[schemes.custom-dark]
foreground = "#d0d0d0"
background = "#1a1a1a"
cursor = "#ff0000"
`
	os.WriteFile(configPath, []byte(config), 0644)

	globalConfigPathOverride = configPath
	defer func() { globalConfigPathOverride = "" }()

	var scripts []string
	setupRunAppleScript = func(script string) error {
		scripts = append(scripts, script)
		return nil
	}
	defer func() { setupRunAppleScript = nil }()

	_, stderr, err := executeCommand("setup", "terminal-app")
	if err != nil {
		t.Fatal(err)
	}

	builtins := BuiltinSchemes()
	expectedCount := len(builtins) + 1 // +1 for custom-dark
	if len(scripts) != expectedCount {
		t.Errorf("expected %d scripts, got %d", expectedCount, len(scripts))
	}

	// Find the custom-dark script and verify its colors.
	var found bool
	for _, script := range scripts {
		if strings.Contains(script, `"custom-dark"`) {
			found = true
			// #d0d0d0 → 208*257=53456
			if !strings.Contains(script, "{53456, 53456, 53456}") {
				t.Errorf("expected foreground RGB {53456, 53456, 53456} in script:\n%s", script)
			}
			// #1a1a1a → 26*257=6682
			if !strings.Contains(script, "{6682, 6682, 6682}") {
				t.Errorf("expected background RGB {6682, 6682, 6682} in script:\n%s", script)
			}
			// #ff0000 → 255*257=65535, 0, 0
			if !strings.Contains(script, "{65535, 0, 0}") {
				t.Errorf("expected cursor RGB {65535, 0, 0} in script:\n%s", script)
			}
			break
		}
	}
	if !found {
		t.Error("expected a script for 'custom-dark' scheme")
	}

	if !strings.Contains(stderr, "custom-dark") {
		t.Errorf("expected custom-dark in stderr output, got: %s", stderr)
	}
}
