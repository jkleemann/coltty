package main

import (
	"fmt"
	"os"
	"sort"
)

type AvailableScheme struct {
	Name   string
	Scheme Scheme
	Tag    string
}

func AvailableSchemes(globalCfg *GlobalConfig) []AvailableScheme {
	entries := make(map[string]AvailableScheme)

	for name, scheme := range BuiltinSchemes() {
		entries[name] = AvailableScheme{
			Name:   name,
			Scheme: scheme,
			Tag:    " (built-in)",
		}
	}

	if globalCfg != nil {
		for name, scheme := range globalCfg.Schemes {
			tag := ""
			if _, ok := entries[name]; ok {
				tag = " (override)"
			}
			entries[name] = AvailableScheme{
				Name:   name,
				Scheme: scheme,
				Tag:    tag,
			}
		}
	}

	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)

	schemes := make([]AvailableScheme, 0, len(names))
	for _, name := range names {
		schemes = append(schemes, entries[name])
	}

	return schemes
}

func LookupScheme(name string, globalCfg *GlobalConfig) (Scheme, bool) {
	if globalCfg != nil {
		if scheme, ok := globalCfg.Schemes[name]; ok {
			return scheme, true
		}
	}
	scheme, ok := builtinSchemes[name]
	return scheme, ok
}

func WriteDirSchemeConfig(path, schemeName string, scheme Scheme, inline bool) error {
	var content string
	if inline {
		content = formatInlineConfig(schemeName, scheme)
	} else {
		content = fmt.Sprintf("scheme = %q\n", schemeName)
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func ResolvedFromScheme(path, schemeName string, scheme Scheme) *ResolvedScheme {
	return &ResolvedScheme{
		Foreground:          scheme.Foreground,
		Background:          scheme.Background,
		Cursor:              scheme.Cursor,
		Palette:             scheme.Palette,
		Source:              path,
		SchemeName:          schemeName,
		Bold:                scheme.Bold,
		SelectionForeground: scheme.SelectionForeground,
		SelectionBackground: scheme.SelectionBackground,
		Tab:                 scheme.Tab,
		ItermPreset:         scheme.ItermPreset,
		TerminalAppProfile:  scheme.TerminalAppProfile,
	}
}
