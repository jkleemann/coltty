package adapter

import (
	"strings"
	"testing"
)

func TestTerminalAppDetect(t *testing.T) {
	a := NewTerminalAppAdapter()

	t.Setenv("TERM_PROGRAM", "ghostty")
	if a.Detect() {
		t.Error("expected false for non-Terminal.app")
	}

	t.Setenv("TERM_PROGRAM", "Apple_Terminal")
	if !a.Detect() {
		t.Error("expected true for Apple_Terminal")
	}
}

func TestTerminalAppName(t *testing.T) {
	a := NewTerminalAppAdapter()
	if a.Name() != "terminal.app" {
		t.Errorf("expected name 'terminal.app', got %q", a.Name())
	}
}

func TestTerminalAppApply(t *testing.T) {
	var capturedScript string
	a := &TerminalAppAdapter{
		RunAppleScript: func(script string) error {
			capturedScript = script
			return nil
		},
	}

	scheme := &ResolvedScheme{
		Name:       "calm",
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Cursor:     "#c0caf5",
	}

	if err := a.Apply(scheme); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(capturedScript, `settings set "calm"`) {
		t.Errorf("expected script to reference 'calm' profile, got: %s", capturedScript)
	}
	if !strings.Contains(capturedScript, "tell application") {
		t.Errorf("expected AppleScript tell block, got: %s", capturedScript)
	}
	if !strings.Contains(capturedScript, "set current settings of front window to targetProfile") {
		t.Errorf("expected profile switch in Apply, got: %s", capturedScript)
	}
	if !strings.Contains(capturedScript, "set normal text color of targetProfile to") {
		t.Errorf("expected normal text color setting, got: %s", capturedScript)
	}
	if !strings.Contains(capturedScript, "set background color of targetProfile to") {
		t.Errorf("expected background color setting, got: %s", capturedScript)
	}
}

func TestTerminalAppApplyWithProfileOverride(t *testing.T) {
	var capturedScript string
	a := &TerminalAppAdapter{
		RunAppleScript: func(script string) error {
			capturedScript = script
			return nil
		},
	}

	scheme := &ResolvedScheme{
		Name:       "calm",
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Cursor:     "#c0caf5",
		Extras: map[string]string{
			"terminal_app_profile": "My Custom Profile",
		},
	}

	if err := a.Apply(scheme); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(capturedScript, `settings set "My Custom Profile"`) {
		t.Errorf("expected profile override, got: %s", capturedScript)
	}
	if strings.Contains(capturedScript, `"calm"`) {
		t.Errorf("should use override profile name, not scheme name")
	}
}

func TestTerminalAppApplyNoName(t *testing.T) {
	a := &TerminalAppAdapter{
		RunAppleScript: func(script string) error {
			t.Fatal("should not have been called")
			return nil
		},
	}

	scheme := &ResolvedScheme{}

	err := a.Apply(scheme)
	if err == nil {
		t.Error("expected error when no scheme name is set")
	}
}

func TestBuildTerminalAppApplyScript(t *testing.T) {
	scheme := &ResolvedScheme{
		Foreground: "#f8f8f2",
		Background: "#282a36",
		Cursor:     "#f8f8f2",
	}

	script, err := BuildTerminalAppApplyScript("dracula", scheme)
	if err != nil {
		t.Fatal(err)
	}

	// Should contain profile creation logic.
	if !strings.Contains(script, `"dracula" is not in profileNames`) {
		t.Error("expected profile existence check")
	}
	if !strings.Contains(script, `make new settings set with properties {name:"dracula"}`) {
		t.Error("expected profile creation via make new")
	}
	if !strings.Contains(script, `set targetProfile to settings set "dracula"`) {
		t.Error("expected target profile assignment")
	}

	// Should contain color settings.
	if !strings.Contains(script, "set normal text color of targetProfile to {63736, 63736, 62194}") {
		t.Errorf("expected foreground color, got:\n%s", script)
	}
	if !strings.Contains(script, "set background color of targetProfile to {10280, 10794, 13878}") {
		t.Errorf("expected background color, got:\n%s", script)
	}
	if !strings.Contains(script, "set cursor color of targetProfile to {63736, 63736, 62194}") {
		t.Errorf("expected cursor color, got:\n%s", script)
	}

	// Apply script should switch profile.
	if !strings.Contains(script, "set current settings of front window to targetProfile") {
		t.Error("expected profile switch in apply script")
	}
}

func TestBuildTerminalAppSetupScript(t *testing.T) {
	scheme := &ResolvedScheme{
		Foreground: "#f8f8f2",
		Background: "#282a36",
		Cursor:     "#f8f8f2",
	}

	script, err := BuildTerminalAppSetupScript("dracula", scheme)
	if err != nil {
		t.Fatal(err)
	}

	// Should contain profile creation and colors.
	if !strings.Contains(script, `"dracula" is not in profileNames`) {
		t.Error("expected profile existence check")
	}
	if !strings.Contains(script, "set normal text color of targetProfile to") {
		t.Error("expected color settings in setup script")
	}

	// Setup script should NOT switch profile.
	if strings.Contains(script, "set current settings of front window") {
		t.Error("setup script should NOT switch the active profile")
	}
}

func TestBuildTerminalAppApplyScriptNoColors(t *testing.T) {
	scheme := &ResolvedScheme{}

	script, err := BuildTerminalAppApplyScript("empty", scheme)
	if err != nil {
		t.Fatal(err)
	}

	// Profile should still be created.
	if !strings.Contains(script, `"empty" is not in profileNames`) {
		t.Error("expected profile creation even with no colors")
	}

	// No color lines should be emitted.
	if strings.Contains(script, "normal text color") {
		t.Error("expected no foreground color line")
	}
	if strings.Contains(script, "background color") {
		t.Error("expected no background color line")
	}
	if strings.Contains(script, "cursor color") {
		t.Error("expected no cursor color line")
	}

	// Should still switch (it's an apply script).
	if !strings.Contains(script, "set current settings of front window to targetProfile") {
		t.Error("expected profile switch even with no colors")
	}
}

func TestBuildTerminalAppApplyScriptInvalidColor(t *testing.T) {
	scheme := &ResolvedScheme{
		Foreground: "not-a-color",
	}

	_, err := BuildTerminalAppApplyScript("bad", scheme)
	if err == nil {
		t.Error("expected error for invalid color")
	}
}
