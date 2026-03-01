package adapter

import (
	"bytes"
	"testing"
)

func TestOSCEmitterFullScheme(t *testing.T) {
	var buf bytes.Buffer
	e := OSCEmitter{Writer: &buf}

	scheme := &ResolvedScheme{
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Cursor:     "#c0caf5",
		Palette:    []string{"#15161e", "#f7768e"},
	}

	e.Emit(scheme)

	got := buf.String()
	expected := "\033]10;#c0caf5\033\\" +
		"\033]11;#1a1b26\033\\" +
		"\033]12;#c0caf5\033\\" +
		"\033]4;0;#15161e\033\\" +
		"\033]4;1;#f7768e\033\\"

	if got != expected {
		t.Errorf("OSCEmitter output mismatch\ngot:  %q\nwant: %q", got, expected)
	}
}

func TestOSCEmitterPartialScheme(t *testing.T) {
	var buf bytes.Buffer
	e := OSCEmitter{Writer: &buf}

	scheme := &ResolvedScheme{
		Background: "#1a1b26",
	}

	e.Emit(scheme)

	got := buf.String()
	expected := "\033]11;#1a1b26\033\\"

	if got != expected {
		t.Errorf("OSCEmitter partial output mismatch\ngot:  %q\nwant: %q", got, expected)
	}
}

func TestOSCEmitterEmptyScheme(t *testing.T) {
	var buf bytes.Buffer
	e := OSCEmitter{Writer: &buf}

	e.Emit(&ResolvedScheme{})

	if buf.Len() != 0 {
		t.Errorf("expected no output for empty scheme, got %q", buf.String())
	}
}

func TestOSCAdapterApply(t *testing.T) {
	var buf bytes.Buffer
	a := &OSCAdapter{
		TermName:   "test-term",
		DetectFunc: func() bool { return true },
		Emitter:    OSCEmitter{Writer: &buf},
	}

	scheme := &ResolvedScheme{
		Foreground: "#ffffff",
		Background: "#000000",
	}

	if err := a.Apply(scheme); err != nil {
		t.Fatal(err)
	}

	if buf.Len() == 0 {
		t.Error("expected OSC output from Apply")
	}
}

func TestOSCAdapterDetectAndName(t *testing.T) {
	a := &OSCAdapter{
		TermName:   "my-terminal",
		DetectFunc: func() bool { return false },
	}

	if a.Name() != "my-terminal" {
		t.Errorf("expected name 'my-terminal', got %q", a.Name())
	}
	if a.Detect() {
		t.Error("expected Detect() false")
	}

	a.DetectFunc = func() bool { return true }
	if !a.Detect() {
		t.Error("expected Detect() true")
	}
}
