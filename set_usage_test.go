package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanThemeUsageCountsNamedSchemes(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "a"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "b"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", ".coltty.toml"), []byte(`scheme = "dracula"`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "b", ".coltty.toml"), []byte(`scheme = "dracula"`), 0644); err != nil {
		t.Fatal(err)
	}

	counts, err := ScanThemeUsage(root)
	if err != nil {
		t.Fatal(err)
	}
	if counts["dracula"] != 2 {
		t.Fatalf("expected dracula count 2, got %d", counts["dracula"])
	}
}

func TestScanThemeUsageIgnoresInlineOnlyConfigs(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "inline"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "inline", ".coltty.toml"), []byte("[overrides]\nbackground = \"#111111\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	counts, err := ScanThemeUsage(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(counts) != 0 {
		t.Fatalf("expected no named counts, got %v", counts)
	}
}
