package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// base16Theme represents a base16 YAML theme file.
type base16Theme struct {
	Scheme string `yaml:"scheme"`
	Author string `yaml:"author"`
	Base00 string `yaml:"base00"`
	Base01 string `yaml:"base01"`
	Base02 string `yaml:"base02"`
	Base03 string `yaml:"base03"`
	Base04 string `yaml:"base04"`
	Base05 string `yaml:"base05"`
	Base06 string `yaml:"base06"`
	Base07 string `yaml:"base07"`
	Base08 string `yaml:"base08"`
	Base09 string `yaml:"base09"`
	Base0A string `yaml:"base0A"`
	Base0B string `yaml:"base0B"`
	Base0C string `yaml:"base0C"`
	Base0D string `yaml:"base0D"`
	Base0E string `yaml:"base0E"`
	Base0F string `yaml:"base0F"`
}

func importBase16(path string) (Scheme, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scheme{}, "", fmt.Errorf("reading file: %w", err)
	}

	var theme base16Theme
	if err := yaml.Unmarshal(data, &theme); err != nil {
		return Scheme{}, "", fmt.Errorf("parsing base16 YAML: %w", err)
	}

	hex := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			return s
		}
		if !strings.HasPrefix(s, "#") {
			s = "#" + s
		}
		return strings.ToLower(s)
	}

	// base16 mapping to 16-color ANSI palette:
	//   palette[0]  = base00 (black)        palette[8]  = base03 (bright black)
	//   palette[1]  = base08 (red)          palette[9]  = base09 (bright red)
	//   palette[2]  = base0B (green)        palette[10] = base0B (bright green)
	//   palette[3]  = base0A (yellow)       palette[11] = base0A (bright yellow)
	//   palette[4]  = base0D (blue)         palette[12] = base0D (bright blue)
	//   palette[5]  = base0E (magenta)      palette[13] = base0E (bright magenta)
	//   palette[6]  = base0C (cyan)         palette[14] = base0C (bright cyan)
	//   palette[7]  = base05 (white)        palette[15] = base07 (bright white)
	palette := []string{
		hex(theme.Base00), hex(theme.Base08), hex(theme.Base0B), hex(theme.Base0A),
		hex(theme.Base0D), hex(theme.Base0E), hex(theme.Base0C), hex(theme.Base05),
		hex(theme.Base03), hex(theme.Base09), hex(theme.Base0B), hex(theme.Base0A),
		hex(theme.Base0D), hex(theme.Base0E), hex(theme.Base0C), hex(theme.Base07),
	}

	scheme := Scheme{
		Foreground: hex(theme.Base05),
		Background: hex(theme.Base00),
		Cursor:     hex(theme.Base05),
		Palette:    palette,
	}

	return scheme, theme.Scheme, nil
}
