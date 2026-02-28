package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGhosttyDetect(t *testing.T) {
	g := NewGhosttyAdapter("")

	// Without TERM_PROGRAM set to ghostty
	t.Setenv("TERM_PROGRAM", "iterm2")
	if g.Detect() {
		t.Error("expected Detect() false for iterm2")
	}

	// With TERM_PROGRAM set to ghostty
	t.Setenv("TERM_PROGRAM", "ghostty")
	if !g.Detect() {
		t.Error("expected Detect() true for ghostty")
	}
}

func TestGhosttyApply(t *testing.T) {
	dir := t.TempDir()
	fragmentPath := filepath.Join(dir, "coltty", "ghostty-colors")

	g := NewGhosttyAdapter(fragmentPath)

	scheme := &ResolvedScheme{
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Cursor:     "#c0caf5",
		Palette: []string{
			"#15161e", "#f7768e", "#9ece6a", "#e0af68",
			"#7aa2f7", "#bb9af7", "#7dcfff", "#a9b1d6",
			"#414868", "#f7768e", "#9ece6a", "#e0af68",
			"#7aa2f7", "#bb9af7", "#7dcfff", "#c0caf5",
		},
	}

	if err := g.Apply(scheme); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(fragmentPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	// Check key lines
	expectedLines := []string{
		"foreground = #c0caf5",
		"background = #1a1b26",
		"cursor-color = #c0caf5",
		"palette = 0=#15161e",
		"palette = 1=#f7768e",
		"palette = 15=#c0caf5",
	}

	for _, line := range expectedLines {
		if !strings.Contains(content, line) {
			t.Errorf("expected fragment to contain %q, got:\n%s", line, content)
		}
	}
}

func TestGhosttyApplyCreatesDir(t *testing.T) {
	dir := t.TempDir()
	fragmentPath := filepath.Join(dir, "nested", "deep", "ghostty-colors")

	g := NewGhosttyAdapter(fragmentPath)

	scheme := &ResolvedScheme{
		Foreground: "#ffffff",
		Background: "#000000",
	}

	if err := g.Apply(scheme); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(fragmentPath); os.IsNotExist(err) {
		t.Error("expected fragment file to be created")
	}
}

func TestGhosttyName(t *testing.T) {
	g := NewGhosttyAdapter("")
	if g.Name() != "ghostty" {
		t.Errorf("expected name 'ghostty', got %q", g.Name())
	}
}

func TestDetectAdapter(t *testing.T) {
	dir := t.TempDir()

	adapters := []TerminalAdapter{
		NewGhosttyAdapter(filepath.Join(dir, "ghostty-colors")),
	}

	// No match
	t.Setenv("TERM_PROGRAM", "xterm")
	a := DetectAdapter(adapters)
	if a != nil {
		t.Error("expected nil adapter for xterm")
	}

	// Ghostty match
	t.Setenv("TERM_PROGRAM", "ghostty")
	a = DetectAdapter(adapters)
	if a == nil {
		t.Fatal("expected ghostty adapter")
	}
	if a.Name() != "ghostty" {
		t.Errorf("expected adapter name 'ghostty', got %q", a.Name())
	}
}
