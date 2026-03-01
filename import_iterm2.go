package main

import (
	"fmt"
	"math"
	"os"

	"howett.net/plist"
)

// itermColorComponent represents an iTerm2 color entry in a plist.
type itermColorComponent struct {
	Red   float64 `plist:"Red Component"`
	Green float64 `plist:"Green Component"`
	Blue  float64 `plist:"Blue Component"`
}

func (c itermColorComponent) toHex() string {
	r := int(math.Round(c.Red * 255))
	g := int(math.Round(c.Green * 255))
	b := int(math.Round(c.Blue * 255))
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func importITerm2(path string) (Scheme, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scheme{}, "", fmt.Errorf("reading file: %w", err)
	}

	var raw map[string]itermColorComponent
	_, err = plist.Unmarshal(data, &raw)
	if err != nil {
		return Scheme{}, "", fmt.Errorf("parsing iTerm2 plist: %w", err)
	}

	getColor := func(key string) string {
		if c, ok := raw[key]; ok {
			return c.toHex()
		}
		return ""
	}

	palette := make([]string, 16)
	for i := 0; i < 16; i++ {
		key := fmt.Sprintf("Ansi %d Color", i)
		palette[i] = getColor(key)
	}

	scheme := Scheme{
		Foreground:          getColor("Foreground Color"),
		Background:          getColor("Background Color"),
		Cursor:              getColor("Cursor Color"),
		Palette:             palette,
		Bold:                getColor("Bold Color"),
		SelectionBackground: getColor("Selection Color"),
		SelectionForeground: getColor("Selected Text Color"),
	}

	return scheme, "", nil
}
