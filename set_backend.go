package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
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

func InferClosestScheme(overrides Scheme, globalCfg *GlobalConfig) (string, bool) {
	candidates := AvailableSchemes(globalCfg)
	if len(candidates) == 0 {
		return "", false
	}

	bestName := ""
	bestScore := -1
	for _, candidate := range candidates {
		score := schemeSimilarityScore(overrides, candidate.Scheme)
		if score > bestScore {
			bestScore = score
			bestName = candidate.Name
		}
	}
	if bestName == "" {
		return "", false
	}
	return bestName, true
}

func schemeSimilarityScore(a, b Scheme) int {
	score := 0
	if a.Foreground != "" {
		score += colorSimilarityScore(a.Foreground, b.Foreground)
	}
	if a.Background != "" {
		score += colorSimilarityScore(a.Background, b.Background)
	}
	if a.Cursor != "" {
		score += colorSimilarityScore(a.Cursor, b.Cursor)
	}
	for i := 0; i < len(a.Palette) && i < len(b.Palette); i++ {
		score += colorSimilarityScore(a.Palette[i], b.Palette[i])
	}
	return score
}

func colorSimilarityScore(a, b string) int {
	if strings.EqualFold(a, b) {
		return 1000
	}

	ar, ag, ab, okA := parseHexColor(a)
	br, bg, bb, okB := parseHexColor(b)
	if !okA || !okB {
		return 0
	}

	diff := absInt(ar-br) + absInt(ag-bg) + absInt(ab-bb)
	return 765 - diff
}

func parseHexColor(value string) (int, int, int, bool) {
	if len(value) != 7 || value[0] != '#' {
		return 0, 0, 0, false
	}

	r, err := strconv.ParseInt(value[1:3], 16, 0)
	if err != nil {
		return 0, 0, 0, false
	}
	g, err := strconv.ParseInt(value[3:5], 16, 0)
	if err != nil {
		return 0, 0, 0, false
	}
	b, err := strconv.ParseInt(value[5:7], 16, 0)
	if err != nil {
		return 0, 0, 0, false
	}
	return int(r), int(g), int(b), true
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
