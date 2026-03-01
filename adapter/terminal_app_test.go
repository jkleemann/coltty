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
		Name: "calm",
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
		Name: "calm",
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
