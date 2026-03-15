package main

import "github.com/charmbracelet/lipgloss"

type previewStyleRoles struct {
	Base     lipgloss.Style
	Muted    lipgloss.Style
	Heading  lipgloss.Style
	Keyword  lipgloss.Style
	Function lipgloss.Style
	String   lipgloss.Style
	Accent   lipgloss.Style
	Bullet   lipgloss.Style
}

func newPreviewStyleRoles(scheme Scheme) previewStyleRoles {
	base := fallbackColor(scheme.Foreground, "#dddddd")
	muted := fallbackColor(pickPaletteColor(scheme.Palette, 8), fallbackColor(scheme.Background, base))

	return previewStyleRoles{
		Base:     lipgloss.NewStyle().Foreground(lipgloss.Color(base)),
		Muted:    lipgloss.NewStyle().Foreground(lipgloss.Color(muted)),
		Heading:  lipgloss.NewStyle().Foreground(lipgloss.Color(fallbackColor(pickPaletteColor(scheme.Palette, 4), base))).Bold(true),
		Keyword:  lipgloss.NewStyle().Foreground(lipgloss.Color(fallbackColor(pickPaletteColor(scheme.Palette, 5), base))).Bold(true),
		Function: lipgloss.NewStyle().Foreground(lipgloss.Color(fallbackColor(pickPaletteColor(scheme.Palette, 6), base))),
		String:   lipgloss.NewStyle().Foreground(lipgloss.Color(fallbackColor(pickPaletteColor(scheme.Palette, 2), base))),
		Accent:   lipgloss.NewStyle().Foreground(lipgloss.Color(fallbackColor(pickPaletteColor(scheme.Palette, 3), base))),
		Bullet:   lipgloss.NewStyle().Foreground(lipgloss.Color(fallbackColor(pickPaletteColor(scheme.Palette, 1), base))).Bold(true),
	}
}

func pickPaletteColor(palette []string, index int) string {
	if index < 0 || index >= len(palette) {
		return ""
	}
	return palette[index]
}

func fallbackColor(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}
