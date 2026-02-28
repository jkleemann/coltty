package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GhosttyAdapter applies color schemes by writing a Ghostty config fragment file
// and emitting OSC escape sequences to update the active terminal immediately.
type GhosttyAdapter struct {
	// FragmentPath is where the Ghostty color fragment is written.
	// If empty, defaults to ~/.config/coltty/ghostty-colors.
	FragmentPath string
}

// NewGhosttyAdapter creates a GhosttyAdapter with the given fragment path.
// If path is empty, defaults to ~/.config/coltty/ghostty-colors.
func NewGhosttyAdapter(fragmentPath string) *GhosttyAdapter {
	if fragmentPath == "" {
		home, _ := os.UserHomeDir()
		fragmentPath = filepath.Join(home, ".config", "coltty", "ghostty-colors")
	}
	return &GhosttyAdapter{FragmentPath: fragmentPath}
}

func (g *GhosttyAdapter) Name() string {
	return "ghostty"
}

func (g *GhosttyAdapter) Detect() bool {
	return os.Getenv("TERM_PROGRAM") == "ghostty"
}

func (g *GhosttyAdapter) Apply(scheme *ResolvedScheme) error {
	// Write config fragment for new windows/tabs
	if err := g.writeFragment(scheme); err != nil {
		return err
	}

	// Emit OSC sequences to update the active terminal immediately
	g.emitOSC(scheme)

	return nil
}

func (g *GhosttyAdapter) writeFragment(scheme *ResolvedScheme) error {
	content := g.renderFragment(scheme)

	dir := filepath.Dir(g.FragmentPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(g.FragmentPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing ghostty fragment: %w", err)
	}

	return nil
}

// emitOSC writes OSC escape sequences to stdout to change terminal colors
// immediately in the current session.
//
//   - OSC 10: set foreground
//   - OSC 11: set background
//   - OSC 12: set cursor color
//   - OSC 4:  set palette color
func (g *GhosttyAdapter) emitOSC(scheme *ResolvedScheme) {
	var b strings.Builder

	if scheme.Foreground != "" {
		fmt.Fprintf(&b, "\033]10;%s\033\\", scheme.Foreground)
	}
	if scheme.Background != "" {
		fmt.Fprintf(&b, "\033]11;%s\033\\", scheme.Background)
	}
	if scheme.Cursor != "" {
		fmt.Fprintf(&b, "\033]12;%s\033\\", scheme.Cursor)
	}
	for i, color := range scheme.Palette {
		fmt.Fprintf(&b, "\033]4;%d;%s\033\\", i, color)
	}

	fmt.Fprint(os.Stdout, b.String())
}

func (g *GhosttyAdapter) renderFragment(scheme *ResolvedScheme) string {
	var b strings.Builder

	if scheme.Foreground != "" {
		fmt.Fprintf(&b, "foreground = %s\n", scheme.Foreground)
	}
	if scheme.Background != "" {
		fmt.Fprintf(&b, "background = %s\n", scheme.Background)
	}
	if scheme.Cursor != "" {
		fmt.Fprintf(&b, "cursor-color = %s\n", scheme.Cursor)
	}
	for i, color := range scheme.Palette {
		fmt.Fprintf(&b, "palette = %d=%s\n", i, color)
	}

	return b.String()
}
