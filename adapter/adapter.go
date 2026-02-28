package adapter

// ResolvedScheme holds the final color values to apply.
type ResolvedScheme struct {
	Foreground string
	Background string
	Cursor     string
	Palette    []string
}

// TerminalAdapter applies color schemes to a specific terminal emulator.
type TerminalAdapter interface {
	// Apply writes the resolved scheme to the terminal.
	Apply(scheme *ResolvedScheme) error
	// Detect returns true if this adapter's terminal is active.
	Detect() bool
	// Name returns a human-readable name for this adapter.
	Name() string
}

// DetectAdapter returns the first adapter whose Detect() returns true,
// or nil if no adapter matches.
func DetectAdapter(adapters []TerminalAdapter) TerminalAdapter {
	for _, a := range adapters {
		if a.Detect() {
			return a
		}
	}
	return nil
}

// AllAdapters returns all available terminal adapters.
func AllAdapters() []TerminalAdapter {
	return []TerminalAdapter{
		NewGhosttyAdapter(""),
	}
}
