package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const dirConfigFile = ".coltty.toml"

// Resolve walks from startDir up to the filesystem root looking for .coltty.toml files.
// The first one found wins. Its scheme is resolved against the global config.
// If none is found, the global default is used.
func Resolve(startDir string, globalCfg *GlobalConfig) (*ResolvedScheme, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute path: %w", err)
	}

	var scheme *ResolvedScheme

	for {
		configPath := filepath.Join(dir, dirConfigFile)
		if _, err := os.Stat(configPath); err == nil {
			dirCfg, err := LoadDirConfig(configPath)
			if err != nil {
				// Config parse error: warn and fall back to default
				fmt.Fprintf(os.Stderr, "coltty: warning: failed to parse %s: %v\n", configPath, err)
				scheme = ResolveScheme(nil, globalCfg, "")
				break
			}
			scheme = ResolveScheme(dirCfg, globalCfg, configPath)
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	if scheme == nil {
		// No .coltty.toml found anywhere; use default
		scheme = ResolveScheme(nil, globalCfg, "")
	}

	if err := scheme.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "coltty: warning: %v\n", err)
	}

	return scheme, nil
}
