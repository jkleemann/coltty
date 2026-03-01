package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Scheme defines a color scheme with foreground, background, cursor, and palette colors.
type Scheme struct {
	Foreground string   `toml:"foreground"`
	Background string   `toml:"background"`
	Cursor     string   `toml:"cursor"`
	Palette    []string `toml:"palette"`

	// Extended colors for terminals that support them (e.g. iTerm2).
	Bold                string `toml:"bold"`
	SelectionForeground string `toml:"selection_foreground"`
	SelectionBackground string `toml:"selection_background"`
	Tab                 string `toml:"tab"`

	// Terminal-specific profile/preset names.
	ItermPreset        string `toml:"iterm_preset"`
	TerminalAppProfile string `toml:"terminal_app_profile"`
}

// GlobalConfig is the top-level config at ~/.config/coltty/config.toml.
type GlobalConfig struct {
	Default struct {
		Scheme string `toml:"scheme"`
	} `toml:"default"`
	Schemes map[string]Scheme `toml:"schemes"`
}

// DirConfig is a per-directory .coltty.toml file.
type DirConfig struct {
	Scheme    string `toml:"scheme"`
	Overrides Scheme `toml:"overrides"`
}

// ResolvedScheme is the final resolved color scheme ready to apply.
type ResolvedScheme struct {
	Foreground string
	Background string
	Cursor     string
	Palette    []string
	Source     string // path to the config file that provided this scheme
	SchemeName string // the resolved scheme name, for profile-based adapters

	// Extended colors.
	Bold                string
	SelectionForeground string
	SelectionBackground string
	Tab                 string
	ItermPreset         string
	TerminalAppProfile  string
}

// builtinSchemes are shipped with the binary and available without any config file.
// User-defined schemes with the same name override these.
var builtinSchemes = map[string]Scheme{
	"gruvbox": {
		Foreground: "#ebdbb2",
		Background: "#282828",
		Cursor:     "#ebdbb2",
		Palette: []string{
			"#282828", "#cc241d", "#98971a", "#d79921",
			"#458588", "#b16286", "#689d6a", "#a89984",
			"#928374", "#fb4934", "#b8bb26", "#fabd2f",
			"#83a598", "#d3869b", "#8ec07c", "#ebdbb2",
		},
	},
	"nord": {
		Foreground: "#d8dee9",
		Background: "#2e3440",
		Cursor:     "#d8dee9",
		Palette: []string{
			"#3b4252", "#bf616a", "#a3be8c", "#ebcb8b",
			"#81a1c1", "#b48ead", "#88c0d0", "#e5e9f0",
			"#4c566a", "#bf616a", "#a3be8c", "#ebcb8b",
			"#81a1c1", "#b48ead", "#8fbcbb", "#eceff4",
		},
	},
	"dracula": {
		Foreground: "#f8f8f2",
		Background: "#282a36",
		Cursor:     "#f8f8f2",
		Palette: []string{
			"#21222c", "#ff5555", "#50fa7b", "#f1fa8c",
			"#bd93f9", "#ff79c6", "#8be9fd", "#f8f8f2",
			"#6272a4", "#ff6e6e", "#69ff94", "#ffffa5",
			"#d6acff", "#ff92df", "#a4ffff", "#ffffff",
		},
	},
	"solarized-dark": {
		Foreground: "#839496",
		Background: "#002b36",
		Cursor:     "#839496",
		Palette: []string{
			"#073642", "#dc322f", "#859900", "#b58900",
			"#268bd2", "#d33682", "#2aa198", "#eee8d5",
			"#002b36", "#cb4b16", "#586e75", "#657b83",
			"#839496", "#6c71c4", "#93a1a1", "#fdf6e3",
		},
	},
	"catppuccin": {
		Foreground: "#cdd6f4",
		Background: "#1e1e2e",
		Cursor:     "#f5e0dc",
		Palette: []string{
			"#45475a", "#f38ba8", "#a6e3a1", "#f9e2af",
			"#89b4fa", "#f5c2e7", "#94e2d5", "#bac2de",
			"#585b70", "#f38ba8", "#a6e3a1", "#f9e2af",
			"#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8",
		},
	},
	"one-dark": {
		Foreground: "#abb2bf",
		Background: "#282c34",
		Cursor:     "#528bff",
		Palette: []string{
			"#282c34", "#e06c75", "#98c379", "#e5c07b",
			"#61afef", "#c678dd", "#56b6c2", "#abb2bf",
			"#545862", "#e06c75", "#98c379", "#e5c07b",
			"#61afef", "#c678dd", "#56b6c2", "#c8ccd4",
		},
	},
	"rose-pine": {
		Foreground: "#e0def4",
		Background: "#191724",
		Cursor:     "#524f67",
		Palette: []string{
			"#26233a", "#eb6f92", "#31748f", "#f6c177",
			"#9ccfd8", "#c4a7e7", "#ebbcba", "#e0def4",
			"#6e6a86", "#eb6f92", "#31748f", "#f6c177",
			"#9ccfd8", "#c4a7e7", "#ebbcba", "#e0def4",
		},
	},
	"kanagawa": {
		Foreground: "#dcd7ba",
		Background: "#1f1f28",
		Cursor:     "#c8c093",
		Palette: []string{
			"#16161d", "#c34043", "#76946a", "#c0a36e",
			"#7e9cd8", "#957fb8", "#6a9589", "#c8c093",
			"#727169", "#e82424", "#98bb6c", "#e6c384",
			"#7fb4ca", "#938aa9", "#7aa89f", "#dcd7ba",
		},
	},
}

// hardcodedDefault is used when no config exists and no built-in is referenced.
var hardcodedDefault = builtinSchemes["gruvbox"]

// BuiltinSchemes returns a copy of all built-in schemes.
func BuiltinSchemes() map[string]Scheme {
	result := make(map[string]Scheme, len(builtinSchemes))
	for k, v := range builtinSchemes {
		result[k] = v
	}
	return result
}

// globalConfigPathOverride is used for testing. If non-empty, it overrides the default path.
var globalConfigPathOverride string

// globalConfigPath returns the path to the global config file.
func globalConfigPath() string {
	if globalConfigPathOverride != "" {
		return globalConfigPathOverride
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "coltty", "config.toml")
}

// LoadGlobalConfig reads and parses the global config file.
// Returns nil with no error if the file doesn't exist.
func LoadGlobalConfig() (*GlobalConfig, error) {
	return LoadGlobalConfigFrom(globalConfigPath())
}

// LoadGlobalConfigFrom reads and parses a global config from a specific path.
func LoadGlobalConfigFrom(path string) (*GlobalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cfg GlobalConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadDirConfig reads and parses a per-directory .coltty.toml file.
func LoadDirConfig(path string) (*DirConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg DirConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ResolveScheme takes a DirConfig and a GlobalConfig and produces a ResolvedScheme.
// If dirCfg is nil, the global default is used.
func ResolveScheme(dirCfg *DirConfig, globalCfg *GlobalConfig, source string) *ResolvedScheme {
	var base Scheme
	var schemeName string

	if dirCfg == nil {
		// No per-directory config; use global default
		base = getDefaultScheme(globalCfg)
		if globalCfg != nil {
			schemeName = globalCfg.Default.Scheme
		}
		if source == "" {
			source = "(default)"
		}
	} else {
		schemeName = dirCfg.Scheme
		if dirCfg.Scheme != "" {
			if globalCfg != nil {
				if s, ok := globalCfg.Schemes[dirCfg.Scheme]; ok {
					base = s
				}
			}
			// Fall back to built-in schemes
			if base.Foreground == "" {
				if s, ok := builtinSchemes[dirCfg.Scheme]; ok {
					base = s
				}
			}
		}
		// Apply overrides
		base = applyOverrides(base, dirCfg.Overrides)
	}

	return &ResolvedScheme{
		Foreground:          base.Foreground,
		Background:          base.Background,
		Cursor:              base.Cursor,
		Palette:             base.Palette,
		Source:              source,
		SchemeName:          schemeName,
		Bold:                base.Bold,
		SelectionForeground: base.SelectionForeground,
		SelectionBackground: base.SelectionBackground,
		Tab:                 base.Tab,
		ItermPreset:         base.ItermPreset,
		TerminalAppProfile:  base.TerminalAppProfile,
	}
}

func getDefaultScheme(globalCfg *GlobalConfig) Scheme {
	if globalCfg == nil {
		return hardcodedDefault
	}
	if globalCfg.Default.Scheme != "" {
		if s, ok := globalCfg.Schemes[globalCfg.Default.Scheme]; ok {
			return s
		}
		if s, ok := builtinSchemes[globalCfg.Default.Scheme]; ok {
			return s
		}
	}
	return hardcodedDefault
}

func applyOverrides(base, overrides Scheme) Scheme {
	if overrides.Foreground != "" {
		base.Foreground = overrides.Foreground
	}
	if overrides.Background != "" {
		base.Background = overrides.Background
	}
	if overrides.Cursor != "" {
		base.Cursor = overrides.Cursor
	}
	if len(overrides.Palette) > 0 {
		base.Palette = overrides.Palette
	}
	if overrides.Bold != "" {
		base.Bold = overrides.Bold
	}
	if overrides.SelectionForeground != "" {
		base.SelectionForeground = overrides.SelectionForeground
	}
	if overrides.SelectionBackground != "" {
		base.SelectionBackground = overrides.SelectionBackground
	}
	if overrides.Tab != "" {
		base.Tab = overrides.Tab
	}
	if overrides.ItermPreset != "" {
		base.ItermPreset = overrides.ItermPreset
	}
	if overrides.TerminalAppProfile != "" {
		base.TerminalAppProfile = overrides.TerminalAppProfile
	}
	return base
}
