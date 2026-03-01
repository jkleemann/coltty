package adapter

import (
	"fmt"
	"os"
	"strings"
)

// ITermAdapter applies color schemes using standard OSC sequences plus
// iTerm2 proprietary OSC 1337 extensions for tab, bold, and selection colors.
type ITermAdapter struct {
	Emitter OSCEmitter
}

// NewITermAdapter creates an ITermAdapter.
func NewITermAdapter() *ITermAdapter {
	return &ITermAdapter{}
}

func (a *ITermAdapter) Name() string {
	return "iterm2"
}

func (a *ITermAdapter) Detect() bool {
	return os.Getenv("TERM_PROGRAM") == "iTerm.app"
}

func (a *ITermAdapter) Apply(scheme *ResolvedScheme) error {
	// Emit standard OSC 10/11/12/4 sequences
	a.Emitter.Emit(scheme)

	// Emit iTerm2 proprietary extensions
	a.emitITermExtras(scheme)

	return nil
}

// emitITermExtras writes iTerm2-specific OSC 1337 sequences for extended colors.
func (a *ITermAdapter) emitITermExtras(scheme *ResolvedScheme) {
	if scheme.Extras == nil {
		return
	}

	w := a.Emitter.Writer
	if w == nil {
		w = os.Stdout
	}

	var b strings.Builder

	if v := scheme.Extras["tab"]; v != "" {
		fmt.Fprintf(&b, "\033]1337;SetColors=tab=%s\033\\", stripHash(v))
	}
	if v := scheme.Extras["bold"]; v != "" {
		fmt.Fprintf(&b, "\033]1337;SetColors=bold=%s\033\\", stripHash(v))
	}
	if v := scheme.Extras["selection_foreground"]; v != "" {
		fmt.Fprintf(&b, "\033]1337;SetColors=selfg=%s\033\\", stripHash(v))
	}
	if v := scheme.Extras["selection_background"]; v != "" {
		fmt.Fprintf(&b, "\033]1337;SetColors=selbg=%s\033\\", stripHash(v))
	}
	if v := scheme.Extras["iterm_preset"]; v != "" {
		fmt.Fprintf(&b, "\033]1337;SetPreset=%s\033\\", v)
	}

	writeOSC(w, b.String())
}

// stripHash removes a leading '#' from a hex color string.
func stripHash(color string) string {
	return strings.TrimPrefix(color, "#")
}
