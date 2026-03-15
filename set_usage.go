package main

import (
	"io/fs"
	"path/filepath"
)

func ScanThemeUsage(root string) (map[string]int, error) {
	counts := make(map[string]int)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != dirConfigFile {
			return nil
		}

		cfg, err := LoadDirConfig(path)
		if err != nil {
			return nil
		}
		if cfg.Scheme == "" {
			return nil
		}
		counts[cfg.Scheme]++
		return nil
	})
	if err != nil {
		return nil, err
	}

	return counts, nil
}
