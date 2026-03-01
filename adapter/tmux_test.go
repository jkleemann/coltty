package adapter

import (
	"bytes"
	"strings"
	"testing"
)

func TestInTmux(t *testing.T) {
	t.Setenv("TMUX", "")
	if InTmux() {
		t.Error("expected false when TMUX is empty")
	}

	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	if !InTmux() {
		t.Error("expected true when TMUX is set")
	}
}

func TestInScreen(t *testing.T) {
	t.Setenv("STY", "")
	if InScreen() {
		t.Error("expected false when STY is empty")
	}

	t.Setenv("STY", "12345.session")
	if !InScreen() {
		t.Error("expected true when STY is set")
	}
}

func TestWrapTmuxPassthrough(t *testing.T) {
	// A single OSC 10 sequence
	input := "\033]10;#c0caf5\033\\"
	wrapped := WrapTmuxPassthrough(input)

	// Should start with DCS: \033Ptmux;
	if !strings.HasPrefix(wrapped, "\033Ptmux;") {
		t.Errorf("expected DCS prefix, got %q", wrapped)
	}

	// Should end with ST: \033\\
	if !strings.HasSuffix(wrapped, "\033\\") {
		t.Errorf("expected ST suffix, got %q", wrapped)
	}

	// ESC characters in the original sequence should be doubled
	// Original: \033]10;#c0caf5\033\\
	// Doubled:  \033\033]10;#c0caf5\033\033\\
	// Wrapped:  \033Ptmux;\033\033]10;#c0caf5\033\033\\\033\\
	expected := "\033Ptmux;\033\033]10;#c0caf5\033\033\\\033\\"
	if wrapped != expected {
		t.Errorf("tmux wrapping mismatch\ngot:  %q\nwant: %q", wrapped, expected)
	}
}

func TestWrapTmuxPassthroughMultiple(t *testing.T) {
	input := "\033]10;#c0caf5\033\\\033]11;#1a1b26\033\\"
	wrapped := WrapTmuxPassthrough(input)

	// Should contain two DCS-wrapped sequences
	count := strings.Count(wrapped, "\033Ptmux;")
	if count != 2 {
		t.Errorf("expected 2 DCS-wrapped sequences, got %d in %q", count, wrapped)
	}
}

func TestSplitOSCSequences(t *testing.T) {
	input := "\033]10;#c0caf5\033\\\033]11;#1a1b26\033\\\033]4;0;#15161e\033\\"
	seqs := splitOSCSequences(input)

	if len(seqs) != 3 {
		t.Fatalf("expected 3 sequences, got %d: %v", len(seqs), seqs)
	}

	if seqs[0] != "\033]10;#c0caf5\033\\" {
		t.Errorf("sequence 0 mismatch: %q", seqs[0])
	}
	if seqs[1] != "\033]11;#1a1b26\033\\" {
		t.Errorf("sequence 1 mismatch: %q", seqs[1])
	}
	if seqs[2] != "\033]4;0;#15161e\033\\" {
		t.Errorf("sequence 2 mismatch: %q", seqs[2])
	}
}

func TestOSCEmitterWithTmux(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")

	var buf bytes.Buffer
	e := OSCEmitter{Writer: &buf}

	scheme := &ResolvedScheme{
		Foreground: "#ffffff",
	}

	e.Emit(scheme)

	got := buf.String()
	// Should be wrapped in DCS passthrough
	if !strings.Contains(got, "\033Ptmux;") {
		t.Errorf("expected tmux DCS wrapping, got %q", got)
	}
}

func TestOSCEmitterWithoutTmux(t *testing.T) {
	t.Setenv("TMUX", "")

	var buf bytes.Buffer
	e := OSCEmitter{Writer: &buf}

	scheme := &ResolvedScheme{
		Foreground: "#ffffff",
	}

	e.Emit(scheme)

	got := buf.String()
	// Should NOT be wrapped
	if strings.Contains(got, "Ptmux") {
		t.Errorf("expected no tmux wrapping, got %q", got)
	}
	expected := "\033]10;#ffffff\033\\"
	if got != expected {
		t.Errorf("output mismatch\ngot:  %q\nwant: %q", got, expected)
	}
}
