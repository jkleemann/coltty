package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// goghTheme represents a Gogh JSON theme file.
type goghTheme struct {
	Name       string `json:"name"`
	Foreground string `json:"foreground"`
	Background string `json:"background"`
	Cursor     string `json:"cursor"`
	Color01    string `json:"color_01"`
	Color02    string `json:"color_02"`
	Color03    string `json:"color_03"`
	Color04    string `json:"color_04"`
	Color05    string `json:"color_05"`
	Color06    string `json:"color_06"`
	Color07    string `json:"color_07"`
	Color08    string `json:"color_08"`
	Color09    string `json:"color_09"`
	Color10    string `json:"color_10"`
	Color11    string `json:"color_11"`
	Color12    string `json:"color_12"`
	Color13    string `json:"color_13"`
	Color14    string `json:"color_14"`
	Color15    string `json:"color_15"`
	Color16    string `json:"color_16"`
}

func importGogh(path string) (Scheme, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scheme{}, "", fmt.Errorf("reading file: %w", err)
	}

	var theme goghTheme
	if err := json.Unmarshal(data, &theme); err != nil {
		return Scheme{}, "", fmt.Errorf("parsing Gogh JSON: %w", err)
	}

	palette := []string{
		normalizeHex(theme.Color01), normalizeHex(theme.Color02),
		normalizeHex(theme.Color03), normalizeHex(theme.Color04),
		normalizeHex(theme.Color05), normalizeHex(theme.Color06),
		normalizeHex(theme.Color07), normalizeHex(theme.Color08),
		normalizeHex(theme.Color09), normalizeHex(theme.Color10),
		normalizeHex(theme.Color11), normalizeHex(theme.Color12),
		normalizeHex(theme.Color13), normalizeHex(theme.Color14),
		normalizeHex(theme.Color15), normalizeHex(theme.Color16),
	}

	scheme := Scheme{
		Foreground: normalizeHex(theme.Foreground),
		Background: normalizeHex(theme.Background),
		Cursor:     normalizeHex(theme.Cursor),
		Palette:    palette,
	}

	return scheme, theme.Name, nil
}

// normalizeHex ensures a color string has a leading # and is lowercase.
func normalizeHex(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if !strings.HasPrefix(s, "#") {
		s = "#" + s
	}
	return strings.ToLower(s)
}
