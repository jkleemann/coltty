package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGlobalConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	content := `
[default]
scheme = "calm"

[schemes.calm]
foreground = "#c0caf5"
background = "#1a1b26"
cursor = "#c0caf5"
palette = [
    "#15161e", "#f7768e", "#9ece6a", "#e0af68",
    "#7aa2f7", "#bb9af7", "#7dcfff", "#a9b1d6",
    "#414868", "#f7768e", "#9ece6a", "#e0af68",
    "#7aa2f7", "#bb9af7", "#7dcfff", "#c0caf5",
]

[schemes.danger]
foreground = "#f8f8f2"
background = "#3b0a0a"
cursor = "#ff5555"
palette = [
    "#282a36", "#ff5555", "#50fa7b", "#f1fa8c",
    "#bd93f9", "#ff79c6", "#8be9fd", "#f8f8f2",
    "#6272a4", "#ff6e6e", "#69ff94", "#ffffa5",
    "#d6acff", "#ff92df", "#a4ffff", "#ffffff",
]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadGlobalConfigFrom(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Default.Scheme != "calm" {
		t.Errorf("expected default scheme 'calm', got %q", cfg.Default.Scheme)
	}

	if len(cfg.Schemes) != 2 {
		t.Errorf("expected 2 schemes, got %d", len(cfg.Schemes))
	}

	calm := cfg.Schemes["calm"]
	if calm.Foreground != "#c0caf5" {
		t.Errorf("expected calm foreground '#c0caf5', got %q", calm.Foreground)
	}
	if len(calm.Palette) != 16 {
		t.Errorf("expected 16 palette colors, got %d", len(calm.Palette))
	}
}

func TestLoadGlobalConfigMissing(t *testing.T) {
	cfg, err := LoadGlobalConfigFrom("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg != nil {
		t.Error("expected nil config for missing file")
	}
}

func TestLoadDirConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".coltty.toml")

	content := `
scheme = "danger"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadDirConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Scheme != "danger" {
		t.Errorf("expected scheme 'danger', got %q", cfg.Scheme)
	}
}

func TestLoadDirConfigWithOverrides(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".coltty.toml")

	content := `
scheme = "calm"

[overrides]
background = "#1e2030"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadDirConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Scheme != "calm" {
		t.Errorf("expected scheme 'calm', got %q", cfg.Scheme)
	}
	if cfg.Overrides.Background != "#1e2030" {
		t.Errorf("expected override background '#1e2030', got %q", cfg.Overrides.Background)
	}
}

func TestFavoritesPathUsesColttyConfigDir(t *testing.T) {
	favoritesConfigPathOverride = filepath.Join(t.TempDir(), "favorites.toml")
	defer func() { favoritesConfigPathOverride = "" }()

	if got := favoritesConfigPath(); got != favoritesConfigPathOverride {
		t.Fatalf("expected favorites override path %q, got %q", favoritesConfigPathOverride, got)
	}
}

func TestLoadFavoritesReturnsEmptyWhenMissing(t *testing.T) {
	favoritesConfigPathOverride = filepath.Join(t.TempDir(), "missing", "favorites.toml")
	defer func() { favoritesConfigPathOverride = "" }()

	cfg, err := LoadFavorites()
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil favorites config")
	}
	if len(cfg.Schemes) != 0 {
		t.Fatalf("expected no favorites, got %v", cfg.Schemes)
	}
}

func TestSaveFavoritesRoundTrip(t *testing.T) {
	favoritesConfigPathOverride = filepath.Join(t.TempDir(), "state", "favorites.toml")
	defer func() { favoritesConfigPathOverride = "" }()

	want := &FavoritesConfig{Schemes: []string{"dracula", "nord"}}
	if err := SaveFavorites(want); err != nil {
		t.Fatal(err)
	}

	got, err := LoadFavorites()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Schemes) != len(want.Schemes) {
		t.Fatalf("expected %d favorites, got %d", len(want.Schemes), len(got.Schemes))
	}
	for i, scheme := range want.Schemes {
		if got.Schemes[i] != scheme {
			t.Fatalf("expected scheme %q at index %d, got %q", scheme, i, got.Schemes[i])
		}
	}
}

func TestResolveSchemeWithDirConfig(t *testing.T) {
	globalCfg := &GlobalConfig{
		Schemes: map[string]Scheme{
			"calm": {
				Foreground: "#c0caf5",
				Background: "#1a1b26",
				Cursor:     "#c0caf5",
				Palette:    []string{"#15161e", "#f7768e"},
			},
		},
	}

	dirCfg := &DirConfig{
		Scheme: "calm",
		Overrides: Scheme{
			Background: "#1e2030",
		},
	}

	resolved := ResolveScheme(dirCfg, globalCfg, "/test/.coltty.toml")

	if resolved.Foreground != "#c0caf5" {
		t.Errorf("expected foreground '#c0caf5', got %q", resolved.Foreground)
	}
	if resolved.Background != "#1e2030" {
		t.Errorf("expected overridden background '#1e2030', got %q", resolved.Background)
	}
	if resolved.Source != "/test/.coltty.toml" {
		t.Errorf("expected source '/test/.coltty.toml', got %q", resolved.Source)
	}
}

func TestResolveSchemeDefault(t *testing.T) {
	resolved := ResolveScheme(nil, nil, "")

	if resolved.Foreground != hardcodedDefault.Foreground {
		t.Errorf("expected default foreground, got %q", resolved.Foreground)
	}
	if resolved.Source != "(default)" {
		t.Errorf("expected source '(default)', got %q", resolved.Source)
	}
}

func TestResolveSchemeGlobalDefault(t *testing.T) {
	globalCfg := &GlobalConfig{
		Schemes: map[string]Scheme{
			"danger": {
				Foreground: "#f8f8f2",
				Background: "#3b0a0a",
				Cursor:     "#ff5555",
			},
		},
	}
	globalCfg.Default.Scheme = "danger"

	resolved := ResolveScheme(nil, globalCfg, "")

	if resolved.Foreground != "#f8f8f2" {
		t.Errorf("expected danger foreground '#f8f8f2', got %q", resolved.Foreground)
	}
}

func TestBuiltinSchemeResolvesWithNoGlobalConfig(t *testing.T) {
	dirCfg := &DirConfig{Scheme: "dracula"}
	resolved := ResolveScheme(dirCfg, nil, "/test/.coltty.toml")

	if resolved.Foreground != "#f8f8f2" {
		t.Errorf("expected dracula foreground '#f8f8f2', got %q", resolved.Foreground)
	}
	if resolved.Background != "#282a36" {
		t.Errorf("expected dracula background '#282a36', got %q", resolved.Background)
	}
}

func TestBuiltinSchemeResolvesWhenGlobalConfigLacksIt(t *testing.T) {
	globalCfg := &GlobalConfig{
		Schemes: map[string]Scheme{
			"custom": {Foreground: "#111", Background: "#222", Cursor: "#333"},
		},
	}

	dirCfg := &DirConfig{Scheme: "nord"}
	resolved := ResolveScheme(dirCfg, globalCfg, "/test/.coltty.toml")

	if resolved.Foreground != "#d8dee9" {
		t.Errorf("expected nord foreground '#d8dee9', got %q", resolved.Foreground)
	}
	if resolved.Background != "#2e3440" {
		t.Errorf("expected nord background '#2e3440', got %q", resolved.Background)
	}
}

func TestUserDefinedSchemeOverridesBuiltin(t *testing.T) {
	globalCfg := &GlobalConfig{
		Schemes: map[string]Scheme{
			"dracula": {
				Foreground: "#custom",
				Background: "#override",
				Cursor:     "#user",
			},
		},
	}

	dirCfg := &DirConfig{Scheme: "dracula"}
	resolved := ResolveScheme(dirCfg, globalCfg, "/test/.coltty.toml")

	if resolved.Foreground != "#custom" {
		t.Errorf("expected user foreground '#custom', got %q", resolved.Foreground)
	}
	if resolved.Background != "#override" {
		t.Errorf("expected user background '#override', got %q", resolved.Background)
	}
}

func TestBuiltinSchemesReturnsAll(t *testing.T) {
	schemes := BuiltinSchemes()

	expected := []string{
		"gruvbox", "nord", "dracula", "solarized-dark",
		"catppuccin", "one-dark", "rose-pine", "kanagawa",
	}

	if len(schemes) != len(expected) {
		t.Errorf("expected %d built-in schemes, got %d", len(expected), len(schemes))
	}

	for _, name := range expected {
		s, ok := schemes[name]
		if !ok {
			t.Errorf("missing built-in scheme %q", name)
			continue
		}
		if s.Foreground == "" || s.Background == "" || s.Cursor == "" {
			t.Errorf("scheme %q has empty required fields", name)
		}
		if len(s.Palette) != 16 {
			t.Errorf("scheme %q has %d palette colors, expected 16", name, len(s.Palette))
		}
	}
}

func TestBuiltinSchemeAsGlobalDefault(t *testing.T) {
	globalCfg := &GlobalConfig{}
	globalCfg.Default.Scheme = "catppuccin"

	resolved := ResolveScheme(nil, globalCfg, "")

	if resolved.Foreground != "#cdd6f4" {
		t.Errorf("expected catppuccin foreground '#cdd6f4', got %q", resolved.Foreground)
	}
	if resolved.Background != "#1e1e2e" {
		t.Errorf("expected catppuccin background '#1e1e2e', got %q", resolved.Background)
	}
}
