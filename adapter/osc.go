package adapter

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// OSCEmitter builds and writes OSC escape sequences to change terminal colors.
type OSCEmitter struct {
	// Writer is the output destination. Defaults to os.Stdout if nil.
	Writer io.Writer
}

// Emit writes OSC 10/11/12/4 escape sequences for the given scheme.
func (e *OSCEmitter) Emit(scheme *ResolvedScheme) {
	w := e.Writer
	if w == nil {
		w = os.Stdout
	}

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

	writeOSC(w, b.String())
}

// writeOSC writes OSC output to the writer, wrapping in tmux DCS passthrough
// if running inside tmux.
func writeOSC(w io.Writer, output string) {
	if output == "" {
		return
	}
	if InTmux() {
		output = WrapTmuxPassthrough(output)
	}
	fmt.Fprint(w, output)
}

// OSCAdapter is a generic adapter for terminals that support standard OSC
// color-setting sequences and need no additional terminal-specific logic.
type OSCAdapter struct {
	TermName   string
	DetectFunc func() bool
	Emitter    OSCEmitter
}

func (a *OSCAdapter) Name() string {
	return a.TermName
}

func (a *OSCAdapter) Detect() bool {
	return a.DetectFunc()
}

func (a *OSCAdapter) Apply(scheme *ResolvedScheme) error {
	a.Emitter.Emit(scheme)
	return nil
}
