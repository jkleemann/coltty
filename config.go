package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
)

// Scheme defines a color scheme with foreground, background, cursor, and palette colors.
type Scheme struct {
	Foreground string   `toml:"foreground"`
	Background string   `toml:"background"`
	Cursor     string   `toml:"cursor"`
	Palette    []string `toml:"palette"`
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
}

// hardcodedDefault is used when no global config exists.
var hardcodedDefault = Scheme{
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

// globalConfigPathOverride is used for testing. If non-empty, it overrides the default path.
var globalConfigPathOverride string

// globalConfigPath returns the path to the global config file.
func globalConfigPath() string {
	if globalConfigPathOverride != "" {
		return globalConfigPathOverride
	}
	return filepath.Join(xdg.ConfigHome, "coltty", "config.toml")
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

	if dirCfg == nil {
		// No per-directory config; use global default
		base = getDefaultScheme(globalCfg)
		if source == "" {
			source = "(default)"
		}
	} else {
		if dirCfg.Scheme != "" && globalCfg != nil {
			if s, ok := globalCfg.Schemes[dirCfg.Scheme]; ok {
				base = s
			}
		}
		// Apply overrides
		base = applyOverrides(base, dirCfg.Overrides)
	}

	return &ResolvedScheme{
		Foreground: base.Foreground,
		Background: base.Background,
		Cursor:     base.Cursor,
		Palette:    base.Palette,
		Source:     source,
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
	return base
}
