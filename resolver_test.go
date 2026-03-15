package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveFindsNearestConfig(t *testing.T) {
	// Create: /tmp/root/.coltty.toml (scheme = "calm")
	//         /tmp/root/child/.coltty.toml (scheme = "danger")
	//         /tmp/root/child/grandchild/

	root := t.TempDir()
	child := filepath.Join(root, "child")
	grandchild := filepath.Join(child, "grandchild")
	os.MkdirAll(grandchild, 0755)

	os.WriteFile(filepath.Join(root, ".coltty.toml"), []byte(`scheme = "calm"`), 0644)
	os.WriteFile(filepath.Join(child, ".coltty.toml"), []byte(`scheme = "danger"`), 0644)

	globalCfg := &GlobalConfig{
		Schemes: map[string]Scheme{
			"calm":   {Foreground: "#aaa", Background: "#111"},
			"danger": {Foreground: "#fff", Background: "#900"},
		},
	}

	// From grandchild, should find child's config (nearest)
	resolved, err := Resolve(grandchild, globalCfg)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Background != "#900" {
		t.Errorf("expected danger background '#900', got %q", resolved.Background)
	}
	if resolved.Source != filepath.Join(child, ".coltty.toml") {
		t.Errorf("expected source in child dir, got %q", resolved.Source)
	}

	// From root, should find root's config
	resolved, err = Resolve(root, globalCfg)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Background != "#111" {
		t.Errorf("expected calm background '#111', got %q", resolved.Background)
	}
}

func TestResolveNoConfig(t *testing.T) {
	dir := t.TempDir()

	resolved, err := Resolve(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	if resolved.Foreground != hardcodedDefault.Foreground {
		t.Errorf("expected hardcoded default, got %q", resolved.Foreground)
	}
	if resolved.Source != "(default)" {
		t.Errorf("expected source '(default)', got %q", resolved.Source)
	}
}

func TestResolveWithOverrides(t *testing.T) {
	dir := t.TempDir()

	config := `
scheme = "calm"

[overrides]
background = "#222222"
`
	os.WriteFile(filepath.Join(dir, ".coltty.toml"), []byte(config), 0644)

	globalCfg := &GlobalConfig{
		Schemes: map[string]Scheme{
			"calm": {
				Foreground: "#c0caf5",
				Background: "#1a1b26",
				Cursor:     "#c0caf5",
			},
		},
	}

	resolved, err := Resolve(dir, globalCfg)
	if err != nil {
		t.Fatal(err)
	}

	if resolved.Background != "#222222" {
		t.Errorf("expected overridden background '#222222', got %q", resolved.Background)
	}
	if resolved.Foreground != "#c0caf5" {
		t.Errorf("expected calm foreground '#c0caf5', got %q", resolved.Foreground)
	}
}

func TestFindDirConfigFindsNearestConfig(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	grandchild := filepath.Join(child, "grandchild")
	if err := os.MkdirAll(grandchild, 0755); err != nil {
		t.Fatal(err)
	}

	wantPath := filepath.Join(child, ".coltty.toml")
	if err := os.WriteFile(wantPath, []byte(`scheme = "dracula"`), 0644); err != nil {
		t.Fatal(err)
	}

	path, cfg, err := FindDirConfig(grandchild)
	if err != nil {
		t.Fatal(err)
	}
	if path != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, path)
	}
	if cfg == nil || cfg.Scheme != "dracula" {
		t.Fatalf("expected dracula config, got %#v", cfg)
	}
}
