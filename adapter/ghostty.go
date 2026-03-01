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
	Emitter      OSCEmitter
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
	g.Emitter.Emit(scheme)

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
