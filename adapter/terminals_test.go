package adapter

import "testing"

func TestAlacrittyDetect(t *testing.T) {
	a := NewAlacrittyAdapter()

	t.Setenv("TERM_PROGRAM", "other")
	if a.Detect() {
		t.Error("expected false for non-Alacritty")
	}

	t.Setenv("TERM_PROGRAM", "Alacritty")
	if !a.Detect() {
		t.Error("expected true for Alacritty")
	}
}

func TestKittyDetect(t *testing.T) {
	a := NewKittyAdapter()

	t.Setenv("TERM_PROGRAM", "other")
	if a.Detect() {
		t.Error("expected false for non-kitty")
	}

	t.Setenv("TERM_PROGRAM", "kitty")
	if !a.Detect() {
		t.Error("expected true for kitty")
	}
}

func TestWezTermDetect(t *testing.T) {
	a := NewWezTermAdapter()

	t.Setenv("TERM_PROGRAM", "other")
	if a.Detect() {
		t.Error("expected false for non-WezTerm")
	}

	t.Setenv("TERM_PROGRAM", "WezTerm")
	if !a.Detect() {
		t.Error("expected true for WezTerm")
	}
}

func TestXtermDetect(t *testing.T) {
	a := NewXtermAdapter()

	t.Setenv("XTERM_VERSION", "")
	if a.Detect() {
		t.Error("expected false when XTERM_VERSION empty")
	}

	t.Setenv("XTERM_VERSION", "XTerm(353)")
	if !a.Detect() {
		t.Error("expected true when XTERM_VERSION set")
	}
}

func TestFootDetect(t *testing.T) {
	a := NewFootAdapter()

	t.Setenv("TERM", "xterm-256color")
	if a.Detect() {
		t.Error("expected false for non-foot TERM")
	}

	t.Setenv("TERM", "foot")
	if !a.Detect() {
		t.Error("expected true for TERM=foot")
	}

	t.Setenv("TERM", "foot-extra")
	if !a.Detect() {
		t.Error("expected true for TERM=foot-extra")
	}
}

func TestKonsoleDetect(t *testing.T) {
	a := NewKonsoleAdapter()

	t.Setenv("KONSOLE_DBUS_SESSION", "")
	if a.Detect() {
		t.Error("expected false when KONSOLE_DBUS_SESSION empty")
	}

	t.Setenv("KONSOLE_DBUS_SESSION", "/Sessions/1")
	if !a.Detect() {
		t.Error("expected true when KONSOLE_DBUS_SESSION set")
	}
}

func TestHyperDetect(t *testing.T) {
	a := NewHyperAdapter()

	t.Setenv("TERM_PROGRAM", "other")
	if a.Detect() {
		t.Error("expected false for non-Hyper")
	}

	t.Setenv("TERM_PROGRAM", "Hyper")
	if !a.Detect() {
		t.Error("expected true for Hyper")
	}
}

func TestTabbyDetect(t *testing.T) {
	a := NewTabbyAdapter()

	t.Setenv("TERM_PROGRAM", "other")
	if a.Detect() {
		t.Error("expected false for non-Tabby")
	}

	t.Setenv("TERM_PROGRAM", "Tabby")
	if !a.Detect() {
		t.Error("expected true for Tabby")
	}
}

func TestStDetect(t *testing.T) {
	a := NewStAdapter()

	t.Setenv("TERM", "xterm-256color")
	if a.Detect() {
		t.Error("expected false for non-st TERM")
	}

	for _, term := range []string{"st-256color", "st", "st-meta-256color", "st-meta"} {
		t.Setenv("TERM", term)
		if !a.Detect() {
			t.Errorf("expected true for TERM=%s", term)
		}
	}
}

func TestUrxvtDetect(t *testing.T) {
	a := NewUrxvtAdapter()

	t.Setenv("TERM", "xterm-256color")
	if a.Detect() {
		t.Error("expected false for non-rxvt TERM")
	}

	for _, term := range []string{"rxvt-unicode-256color", "rxvt-unicode", "rxvt"} {
		t.Setenv("TERM", term)
		if !a.Detect() {
			t.Errorf("expected true for TERM=%s", term)
		}
	}
}

func TestVTEDetect(t *testing.T) {
	a := NewVTEAdapter()

	t.Setenv("VTE_VERSION", "")
	if a.Detect() {
		t.Error("expected false when VTE_VERSION empty")
	}

	t.Setenv("VTE_VERSION", "7201")
	if !a.Detect() {
		t.Error("expected true when VTE_VERSION set")
	}
}

func TestDetectAdapterPriority(t *testing.T) {
	adapters := AllAdapters()

	// Clear all detection env vars to prevent cross-contamination.
	t.Setenv("KONSOLE_DBUS_SESSION", "")
	t.Setenv("VTE_VERSION", "")
	t.Setenv("XTERM_VERSION", "")
	t.Setenv("TERM", "xterm-256color")

	// Ghostty should be detected first when set
	t.Setenv("TERM_PROGRAM", "ghostty")
	a := DetectAdapter(adapters)
	if a == nil || a.Name() != "ghostty" {
		t.Errorf("expected ghostty adapter, got %v", a)
	}

	// Alacritty should be detected
	t.Setenv("TERM_PROGRAM", "Alacritty")
	a = DetectAdapter(adapters)
	if a == nil || a.Name() != "alacritty" {
		t.Errorf("expected alacritty adapter, got %v", a)
	}

	// Konsole should be detected
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KONSOLE_DBUS_SESSION", "/Sessions/1")
	a = DetectAdapter(adapters)
	if a == nil || a.Name() != "konsole" {
		t.Errorf("expected konsole adapter, got %v", a)
	}
	t.Setenv("KONSOLE_DBUS_SESSION", "")

	// VTE should be detected as catch-all
	t.Setenv("VTE_VERSION", "7201")
	a = DetectAdapter(adapters)
	if a == nil || a.Name() != "vte" {
		t.Errorf("expected vte adapter, got %v", a)
	}
	t.Setenv("VTE_VERSION", "")

	// No match should return nil
	t.Setenv("TERM_PROGRAM", "unknown")
	a = DetectAdapter(adapters)
	if a != nil {
		t.Errorf("expected nil for unknown terminal, got %s", a.Name())
	}
}
