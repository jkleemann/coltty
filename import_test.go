package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestImportGogh(t *testing.T) {
	scheme, name, err := importGogh("testdata/dracula.json")
	if err != nil {
		t.Fatal(err)
	}
	if name != "Dracula" {
		t.Errorf("expected name 'Dracula', got %q", name)
	}
	if scheme.Foreground != "#f8f8f2" {
		t.Errorf("expected foreground '#f8f8f2', got %q", scheme.Foreground)
	}
	if scheme.Background != "#282a36" {
		t.Errorf("expected background '#282a36', got %q", scheme.Background)
	}
	if scheme.Cursor != "#f8f8f2" {
		t.Errorf("expected cursor '#f8f8f2', got %q", scheme.Cursor)
	}
	if len(scheme.Palette) != 16 {
		t.Fatalf("expected 16 palette colors, got %d", len(scheme.Palette))
	}
	if scheme.Palette[0] != "#21222c" {
		t.Errorf("expected palette[0] '#21222c', got %q", scheme.Palette[0])
	}
	if scheme.Palette[1] != "#ff5555" {
		t.Errorf("expected palette[1] '#ff5555', got %q", scheme.Palette[1])
	}
}

func TestImportBase16(t *testing.T) {
	scheme, name, err := importBase16("testdata/monokai.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if name != "Monokai" {
		t.Errorf("expected name 'Monokai', got %q", name)
	}
	if scheme.Foreground != "#f8f8f2" {
		t.Errorf("expected foreground '#f8f8f2', got %q", scheme.Foreground)
	}
	if scheme.Background != "#272822" {
		t.Errorf("expected background '#272822', got %q", scheme.Background)
	}
	if len(scheme.Palette) != 16 {
		t.Fatalf("expected 16 palette colors, got %d", len(scheme.Palette))
	}
	// palette[0] = base00
	if scheme.Palette[0] != "#272822" {
		t.Errorf("expected palette[0] '#272822', got %q", scheme.Palette[0])
	}
	// palette[1] = base08 (red)
	if scheme.Palette[1] != "#f92672" {
		t.Errorf("expected palette[1] '#f92672', got %q", scheme.Palette[1])
	}
	// palette[4] = base0D (blue)
	if scheme.Palette[4] != "#66d9ef" {
		t.Errorf("expected palette[4] '#66d9ef', got %q", scheme.Palette[4])
	}
}

func TestImportITerm2(t *testing.T) {
	scheme, _, err := importITerm2("testdata/solarized.itermcolors")
	if err != nil {
		t.Fatal(err)
	}
	if scheme.Foreground == "" {
		t.Error("expected foreground to be set")
	}
	if scheme.Background == "" {
		t.Error("expected background to be set")
	}
	if scheme.Cursor == "" {
		t.Error("expected cursor to be set")
	}
	if len(scheme.Palette) != 16 {
		t.Fatalf("expected 16 palette colors, got %d", len(scheme.Palette))
	}
	for i, c := range scheme.Palette {
		if c == "" {
			t.Errorf("expected palette[%d] to be set", i)
		}
		if !strings.HasPrefix(c, "#") {
			t.Errorf("expected palette[%d] to start with '#', got %q", i, c)
		}
	}
	// iTerm2 extended colors
	if scheme.Bold == "" {
		t.Error("expected bold color to be set")
	}
	if scheme.SelectionBackground == "" {
		t.Error("expected selection background to be set")
	}
	if scheme.SelectionForeground == "" {
		t.Error("expected selection foreground to be set")
	}
}

func TestImportCommandGoghStdout(t *testing.T) {
	stdout, _, err := executeCommand(newTestRootCmd(), "import", "testdata/dracula.json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "[schemes.dracula]") {
		t.Error("expected [schemes.dracula] in TOML output")
	}
	if !strings.Contains(stdout, `foreground = "#f8f8f2"`) {
		t.Error("expected foreground color in output")
	}
	if !strings.Contains(stdout, `"#ff5555"`) {
		t.Error("expected palette colors in output")
	}
}

func TestImportCommandWithName(t *testing.T) {
	stdout, _, err := executeCommand(newTestRootCmd(), "import", "testdata/dracula.json", "--name", "my-theme")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "[schemes.my-theme]") {
		t.Errorf("expected [schemes.my-theme] in output, got:\n%s", stdout)
	}
}

func TestImportCommandBase16(t *testing.T) {
	stdout, _, err := executeCommand(newTestRootCmd(), "import", "testdata/monokai.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "[schemes.monokai]") {
		t.Error("expected [schemes.monokai] in TOML output")
	}
	if !strings.Contains(stdout, `background = "#272822"`) {
		t.Error("expected monokai background in output")
	}
}

func TestImportCommandITerm2(t *testing.T) {
	stdout, _, err := executeCommand(newTestRootCmd(), "import", "testdata/solarized.itermcolors")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "[schemes.solarized]") {
		t.Error("expected [schemes.solarized] in TOML output")
	}
	if !strings.Contains(stdout, "bold = ") {
		t.Error("expected bold color in iTerm2 import")
	}
}

func TestImportCommandAppend(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.toml")

	globalConfigPathOverride = configPath
	defer func() { globalConfigPathOverride = "" }()

	_, stderr, err := executeCommand(newTestRootCmd(), "import", "testdata/dracula.json", "--append")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr, "imported scheme") {
		t.Errorf("expected import confirmation on stderr, got: %s", stderr)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal("expected global config to be written:", err)
	}
	content := string(data)
	if !strings.Contains(content, "dracula") {
		t.Errorf("expected 'dracula' in global config, got:\n%s", content)
	}
}

func TestImportCommandListFormats(t *testing.T) {
	stdout, _, err := executeCommand(newTestRootCmd(), "import", "--list-formats")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "gogh") {
		t.Error("expected 'gogh' in format list")
	}
	if !strings.Contains(stdout, "base16") {
		t.Error("expected 'base16' in format list")
	}
	if !strings.Contains(stdout, "iterm2") {
		t.Error("expected 'iterm2' in format list")
	}
}

func TestImportCommandUnknownFormat(t *testing.T) {
	_, _, err := executeCommand(newTestRootCmd(), "import", "testdata/dracula.json", "--format", "badformat")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestImportCommandAutoDetect(t *testing.T) {
	// .json → gogh
	if f := detectFormat("theme.json"); f != "gogh" {
		t.Errorf("expected 'gogh' for .json, got %q", f)
	}
	// .yaml → base16
	if f := detectFormat("theme.yaml"); f != "base16" {
		t.Errorf("expected 'base16' for .yaml, got %q", f)
	}
	// .yml → base16
	if f := detectFormat("theme.yml"); f != "base16" {
		t.Errorf("expected 'base16' for .yml, got %q", f)
	}
	// .itermcolors → iterm2
	if f := detectFormat("theme.itermcolors"); f != "iterm2" {
		t.Errorf("expected 'iterm2' for .itermcolors, got %q", f)
	}
	// unknown
	if f := detectFormat("theme.txt"); f != "" {
		t.Errorf("expected empty for .txt, got %q", f)
	}
}
func TestAppendToGlobalConfigAtomic(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.toml")

	globalConfigPathOverride = configPath
	defer func() { globalConfigPathOverride = "" }()

	// Seed an initial config with one scheme.
	initial := `[schemes.seed]
foreground = "#111111"
background = "#222222"
cursor = "#333333"
palette = [
    "#000000", "#111111", "#222222", "#333333",
    "#444444", "#555555", "#666666", "#777777",
    "#888888", "#999999", "#aaaaaa", "#bbbbbb",
    "#cccccc", "#dddddd", "#eeeeee", "#ffffff",
]
`

	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Concurrently append many unique schemes.
	const num = 20
	var wg sync.WaitGroup
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("scheme-%02d", idx)
			scheme := Scheme{
				Foreground: "#c0caf5",
				Background: "#1a1b26",
				Cursor:     "#c0caf5",
				Palette: []string{
					"#15161e", "#f7768e", "#9ece6a", "#e0af68",
					"#7aa2f7", "#bb9af7", "#7dcfff", "#a9b1d6",
					"#414868", "#f7768e", "#9ece6a", "#e0af68",
					"#7aa2f7", "#bb9af7", "#7dcfff", "#c0caf5",
				},
			}
			if err := appendToGlobalConfig(name, scheme); err != nil {
				t.Errorf("appendToGlobalConfig failed: %v", err)
			}
		}(i)
	}
	wg.Wait()

	// Load the final config and verify all schemes are present.
	cfg, err := LoadGlobalConfigFrom(configPath)
	if err != nil {
		t.Fatalf("failed to load final config: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if len(cfg.Schemes) != num+1 {
		t.Errorf("expected %d schemes (1 seed + %d appended), got %d", num+1, num, len(cfg.Schemes))
	}

	for i := 0; i < num; i++ {
		name := fmt.Sprintf("scheme-%02d", i)
		if _, ok := cfg.Schemes[name]; !ok {
			t.Errorf("missing scheme %q in final config", name)
		}
	}
	if _, ok := cfg.Schemes["seed"]; !ok {
		t.Error("missing seed scheme in final config")
	}
}

func TestAppendToGlobalConfigOverwriteWarning(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.toml")

	globalConfigPathOverride = configPath
	defer func() { globalConfigPathOverride = "" }()

	// Seed a config with one scheme.
	initial := `[schemes.dracula]
foreground = "#111111"
background = "#222222"
cursor = "#333333"
`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Capture stderr.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	scheme := Scheme{Foreground: "#ffffff", Background: "#000000", Cursor: "#ff0000"}
	err := appendToGlobalConfig("dracula", scheme)

	w.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	stderrOutput := buf.String()
	if !strings.Contains(stderrOutput, "overwriting existing scheme") {
		t.Errorf("expected overwrite warning, got: %q", stderrOutput)
	}
}

func TestNormalizeHexEmpty(t *testing.T) {
	if got := normalizeHex(""); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
	if got := normalizeHex("   "); got != "" {
		t.Errorf("expected empty string for whitespace, got %q", got)
	}
}

func TestNormalizeHexWithHash(t *testing.T) {
	if got := normalizeHex("#FF0000"); got != "#ff0000" {
		t.Errorf("expected '#ff0000', got %q", got)
	}
}

func TestNormalizeHexWithoutHash(t *testing.T) {
	if got := normalizeHex("FF0000"); got != "#ff0000" {
		t.Errorf("expected '#ff0000', got %q", got)
	}
}

func TestImportGoghMissingFile(t *testing.T) {
	_, _, err := importGogh("/nonexistent/path/theme.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestImportGoghInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("not json"), 0644)
	_, _, err := importGogh(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestImportBase16MissingFile(t *testing.T) {
	_, _, err := importBase16("/nonexistent/path/theme.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestImportBase16InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte("not: yaml: ["), 0644)
	_, _, err := importBase16(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestImportITerm2MissingFile(t *testing.T) {
	_, _, err := importITerm2("/nonexistent/path/theme.itermcolors")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestImportITerm2InvalidPlist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.itermcolors")
	os.WriteFile(path, []byte("not a plist"), 0644)
	_, _, err := importITerm2(path)
	if err == nil {
		t.Error("expected error for invalid plist")
	}
}

func TestFormatSchemeTomlEmptyPalette(t *testing.T) {
	scheme := Scheme{
		Foreground: "#c0caf5",
		Background: "#1a1b26",
		Cursor:     "#c0caf5",
	}
	result := formatSchemeToml("minimal", scheme)
	if strings.Contains(result, "palette = [") {
		t.Error("expected no palette section for empty palette")
	}
	if !strings.Contains(result, "[schemes.minimal]") {
		t.Error("expected scheme header")
	}
}

func TestFormatSchemeTomlExtendedColors(t *testing.T) {
	scheme := Scheme{
		Foreground:          "#c0caf5",
		Background:          "#1a1b26",
		Cursor:              "#c0caf5",
		Bold:                "#ffffff",
		SelectionForeground: "#000000",
		SelectionBackground: "#111111",
		Tab:                 "#222222",
	}
	result := formatSchemeToml("extended", scheme)
	if !strings.Contains(result, `bold = "#ffffff"`) {
		t.Error("expected bold color")
	}
	if !strings.Contains(result, `selection_foreground = "#000000"`) {
		t.Error("expected selection_foreground color")
	}
	if !strings.Contains(result, `selection_background = "#111111"`) {
		t.Error("expected selection_background color")
	}
	if !strings.Contains(result, `tab = "#222222"`) {
		t.Error("expected tab color")
	}
}

func TestDeriveSchemeNameCollision(t *testing.T) {
	used := make(map[string]bool)

	name1 := deriveSchemeName("Solarized Dark.json", used)
	if name1 != "solarizeddark" {
		t.Errorf("expected 'solarizeddark', got %q", name1)
	}

	name2 := deriveSchemeName("solarized-dark.json", used)
	if name2 != "solarizeddark-2" {
		t.Errorf("expected 'solarizeddark-2', got %q", name2)
	}

	name3 := deriveSchemeName("Solarized_Dark.yml", used)
	if name3 != "solarizeddark-3" {
		t.Errorf("expected 'solarizeddark-3', got %q", name3)
	}
}
