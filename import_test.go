package main

import (
	"os"
	"path/filepath"
	"strings"
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
	stdout, _, err := executeCommand("import", "testdata/dracula.json")
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
	defer func() { importName = "" }()

	stdout, _, err := executeCommand("import", "testdata/dracula.json", "--name", "my-theme")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "[schemes.my-theme]") {
		t.Errorf("expected [schemes.my-theme] in output, got:\n%s", stdout)
	}
}

func TestImportCommandBase16(t *testing.T) {
	stdout, _, err := executeCommand("import", "testdata/monokai.yaml")
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
	stdout, _, err := executeCommand("import", "testdata/solarized.itermcolors")
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

	defer func() { importAppend = false }()

	_, stderr, err := executeCommand("import", "testdata/dracula.json", "--append")
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
	defer func() { importListFormats = false }()

	stdout, _, err := executeCommand("import", "--list-formats")
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
	defer func() { importFormat = "" }()

	_, _, err := executeCommand("import", "testdata/dracula.json", "--format", "badformat")
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
