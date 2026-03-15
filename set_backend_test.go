package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAvailableSchemesIncludesBuiltinsAndOverrides(t *testing.T) {
	globalCfg := &GlobalConfig{
		Schemes: map[string]Scheme{
			"dracula": {
				Foreground: "#custom",
				Background: "#override",
				Cursor:     "#user",
			},
			"calm": {
				Foreground: "#c0caf5",
				Background: "#1a1b26",
				Cursor:     "#c0caf5",
			},
		},
	}
	globalCfg.Default.Scheme = "calm"

	schemes := AvailableSchemes(globalCfg)
	if len(schemes) == 0 {
		t.Fatal("expected schemes")
	}

	seenCalm := false
	seenDracula := false
	for _, scheme := range schemes {
		switch scheme.Name {
		case "calm":
			seenCalm = true
			if scheme.Tag != "" {
				t.Fatalf("expected calm to have no tag, got %q", scheme.Tag)
			}
		case "dracula":
			seenDracula = true
			if scheme.Tag != " (override)" {
				t.Fatalf("expected dracula override tag, got %q", scheme.Tag)
			}
			if scheme.Scheme.Foreground != "#custom" {
				t.Fatalf("expected override foreground, got %q", scheme.Scheme.Foreground)
			}
		}
	}

	if !seenCalm {
		t.Fatal("expected user-defined calm scheme")
	}
	if !seenDracula {
		t.Fatal("expected overridden dracula scheme")
	}
}

func TestWriteDirSchemeConfigNamed(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".coltty.toml")
	scheme := BuiltinSchemes()["dracula"]

	if err := WriteDirSchemeConfig(path, "dracula", scheme, false); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != `scheme = "dracula"` {
		t.Fatalf("unexpected config: %s", string(data))
	}
}

func TestWriteDirSchemeConfigInline(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".coltty.toml")
	scheme := BuiltinSchemes()["dracula"]

	if err := WriteDirSchemeConfig(path, "dracula", scheme, true); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "[overrides]") {
		t.Fatal("expected overrides block")
	}
	if !strings.Contains(content, `foreground = "#f8f8f2"`) {
		t.Fatal("expected dracula foreground")
	}
}
