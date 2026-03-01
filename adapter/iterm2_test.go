package adapter

import (
	"bytes"
	"strings"
	"testing"
)

func TestITermDetect(t *testing.T) {
	a := NewITermAdapter()

	t.Setenv("TERM_PROGRAM", "ghostty")
	if a.Detect() {
		t.Error("expected false for non-iTerm")
	}

	t.Setenv("TERM_PROGRAM", "iTerm.app")
	if !a.Detect() {
		t.Error("expected true for iTerm.app")
	}
}

func TestITermName(t *testing.T) {
	a := NewITermAdapter()
	if a.Name() != "iterm2" {
		t.Errorf("expected name 'iterm2', got %q", a.Name())
	}
}

func TestITermApplyStandardOnly(t *testing.T) {
	var buf bytes.Buffer
	a := &ITermAdapter{Emitter: OSCEmitter{Writer: &buf}}

	scheme := &ResolvedScheme{
		Foreground: "#c0caf5",
		Background: "#1a1b26",
	}

	if err := a.Apply(scheme); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	// Should contain standard OSC sequences
	if !strings.Contains(got, "\033]10;#c0caf5\033\\") {
		t.Error("expected OSC 10 foreground sequence")
	}
	if !strings.Contains(got, "\033]11;#1a1b26\033\\") {
		t.Error("expected OSC 11 background sequence")
	}
	// Should NOT contain any OSC 1337
	if strings.Contains(got, "1337") {
		t.Error("expected no OSC 1337 sequences without extras")
	}
}

func TestITermApplyWithExtras(t *testing.T) {
	var buf bytes.Buffer
	a := &ITermAdapter{Emitter: OSCEmitter{Writer: &buf}}

	scheme := &ResolvedScheme{
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Extras: map[string]string{
			"tab":                  "#ff0000",
			"bold":                 "#00ff00",
			"selection_foreground": "#ffffff",
			"selection_background": "#333333",
		},
	}

	if err := a.Apply(scheme); err != nil {
		t.Fatal(err)
	}

	got := buf.String()

	// Verify OSC 1337 sequences use stripped hex (no #)
	expectedParts := []string{
		"\033]1337;SetColors=tab=ff0000\033\\",
		"\033]1337;SetColors=bold=00ff00\033\\",
		"\033]1337;SetColors=selfg=ffffff\033\\",
		"\033]1337;SetColors=selbg=333333\033\\",
	}

	for _, part := range expectedParts {
		if !strings.Contains(got, part) {
			t.Errorf("expected output to contain %q\ngot: %q", part, got)
		}
	}
}

func TestITermApplyWithPreset(t *testing.T) {
	var buf bytes.Buffer
	a := &ITermAdapter{Emitter: OSCEmitter{Writer: &buf}}

	scheme := &ResolvedScheme{
		Foreground: "#c0caf5",
		Extras: map[string]string{
			"iterm_preset": "Solarized Dark",
		},
	}

	if err := a.Apply(scheme); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, "\033]1337;SetPreset=Solarized Dark\033\\") {
		t.Errorf("expected SetPreset sequence, got: %q", got)
	}
}
