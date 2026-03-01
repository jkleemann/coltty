package adapter

// ResolvedScheme holds the final color values to apply.
type ResolvedScheme struct {
	Foreground string
	Background string
	Cursor     string
	Palette    []string
	Name       string            // scheme name, used by profile-based adapters
	Extras     map[string]string // terminal-specific extended colors
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
// Order matters: more specific adapters come first.
func AllAdapters() []TerminalAdapter {
	return []TerminalAdapter{
		// macOS-specific (most specific first)
		NewGhosttyAdapter(""),
		NewITermAdapter(),
		NewTerminalAppAdapter(),
		// Cross-platform (TERM_PROGRAM detection)
		NewAlacrittyAdapter(),
		NewKittyAdapter(),
		NewWezTermAdapter(),
		NewHyperAdapter(),
		NewTabbyAdapter(),
		// Linux-specific (env var detection)
		NewKonsoleAdapter(),
		// TERM-based detection (less specific)
		NewXtermAdapter(),
		NewFootAdapter(),
		NewStAdapter(),
		NewUrxvtAdapter(),
		// VTE catch-all (must be last — matches any VTE-based terminal)
		NewVTEAdapter(),
	}
}
