package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var importFormat string
var importName string
var importAppend bool
var importListFormats bool

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a color scheme from Gogh, base16, or iTerm2 format",
	Long: `Import a color scheme from an external theme file.

Supported formats:
  gogh      Gogh JSON theme files (.json)
  base16    base16 YAML theme files (.yaml, .yml)
  iterm2    iTerm2 color preset files (.itermcolors)

Format is auto-detected from the file extension, or set explicitly with --format.
Outputs TOML to stdout by default. Use --append to write directly to the global config.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if importListFormats {
			fmt.Println("Supported import formats:")
			fmt.Println("  gogh      Gogh JSON theme files (.json)")
			fmt.Println("  base16    base16 YAML theme files (.yaml, .yml)")
			fmt.Println("  iterm2    iTerm2 color preset files (.itermcolors)")
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("requires a file argument (use --list-formats to see supported formats)")
		}

		path := args[0]
		format := importFormat
		if format == "" {
			format = detectFormat(path)
			if format == "" {
				return fmt.Errorf("cannot detect format from file extension %q (use --format to specify)", filepath.Ext(path))
			}
		}

		scheme, detectedName, err := importFile(path, format)
		if err != nil {
			return err
		}

		name := importName
		if name == "" {
			if detectedName != "" {
				name = strings.ToLower(strings.ReplaceAll(detectedName, " ", "-"))
			} else {
				// Derive from filename.
				base := filepath.Base(path)
				name = strings.TrimSuffix(base, filepath.Ext(base))
				name = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
			}
		}

		if importAppend {
			return appendToGlobalConfig(name, scheme)
		}

		// Output TOML to stdout.
		fmt.Print(formatSchemeToml(name, scheme))
		return nil
	},
}

func detectFormat(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return "gogh"
	case ".yaml", ".yml":
		return "base16"
	case ".itermcolors":
		return "iterm2"
	default:
		return ""
	}
}

func importFile(path, format string) (Scheme, string, error) {
	switch format {
	case "gogh":
		return importGogh(path)
	case "base16":
		return importBase16(path)
	case "iterm2":
		return importITerm2(path)
	default:
		return Scheme{}, "", fmt.Errorf("unknown format %q (supported: gogh, base16, iterm2)", format)
	}
}

// formatSchemeToml formats a scheme as a TOML snippet suitable for pasting into config.toml.
func formatSchemeToml(name string, s Scheme) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[schemes.%s]\n", name)
	fmt.Fprintf(&b, "foreground = %q\n", s.Foreground)
	fmt.Fprintf(&b, "background = %q\n", s.Background)
	fmt.Fprintf(&b, "cursor = %q\n", s.Cursor)
	if len(s.Palette) > 0 {
		b.WriteString("palette = [\n")
		for i := 0; i < len(s.Palette); i += 4 {
			end := i + 4
			if end > len(s.Palette) {
				end = len(s.Palette)
			}
			quoted := make([]string, end-i)
			for j, c := range s.Palette[i:end] {
				quoted[j] = fmt.Sprintf("%q", c)
			}
			b.WriteString("    ")
			b.WriteString(strings.Join(quoted, ", "))
			if end < len(s.Palette) {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString("]\n")
	}
	if s.Bold != "" {
		fmt.Fprintf(&b, "bold = %q\n", s.Bold)
	}
	if s.SelectionForeground != "" {
		fmt.Fprintf(&b, "selection_foreground = %q\n", s.SelectionForeground)
	}
	if s.SelectionBackground != "" {
		fmt.Fprintf(&b, "selection_background = %q\n", s.SelectionBackground)
	}
	if s.Tab != "" {
		fmt.Fprintf(&b, "tab = %q\n", s.Tab)
	}
	return b.String()
}

// appendToGlobalConfig reads the existing global config, adds the scheme, and writes it back.
func appendToGlobalConfig(name string, scheme Scheme) error {
	configPath := globalConfigPath()

	cfg, err := LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("loading global config: %w", err)
	}
	if cfg == nil {
		cfg = &GlobalConfig{}
	}
	if cfg.Schemes == nil {
		cfg.Schemes = make(map[string]Scheme)
	}

	if _, exists := cfg.Schemes[name]; exists {
		fmt.Fprintf(os.Stderr, "coltty: overwriting existing scheme %q in global config\n", name)
	}
	cfg.Schemes[name] = scheme

	// Ensure the config directory exists.
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("writing global config: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("encoding global config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "coltty: imported scheme %q to %s\n", name, configPath)
	return nil
}

func init() {
	importCmd.Flags().StringVar(&importFormat, "format", "", "theme format: gogh, base16, or iterm2 (auto-detected from extension)")
	importCmd.Flags().StringVar(&importName, "name", "", "scheme name (default: derived from file or theme metadata)")
	importCmd.Flags().BoolVar(&importAppend, "append", false, "write directly to global config instead of stdout")
	importCmd.Flags().BoolVar(&importListFormats, "list-formats", false, "list supported import formats")

	rootCmd.AddCommand(importCmd)
}
