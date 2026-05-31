package adapter

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidateHex checks whether color is a valid 6-digit hex color like "#rrggbb".
// It returns an error describing why the color is invalid.
func ValidateHex(color string) error {
	hex := strings.TrimPrefix(color, "#")
	if len(hex) != 6 {
		return fmt.Errorf("invalid hex color %q: expected 6 hex digits", color)
	}
	if _, err := strconv.ParseUint(hex[0:2], 16, 8); err != nil {
		return fmt.Errorf("invalid hex color %q: %w", color, err)
	}
	if _, err := strconv.ParseUint(hex[2:4], 16, 8); err != nil {
		return fmt.Errorf("invalid hex color %q: %w", color, err)
	}
	if _, err := strconv.ParseUint(hex[4:6], 16, 8); err != nil {
		return fmt.Errorf("invalid hex color %q: %w", color, err)
	}
	return nil
}

// HexToTerminalAppRGB converts a hex color string like "#f8f8f2" to a
// Terminal.app AppleScript RGB list like "{63736, 63736, 62194}".
// Terminal.app uses 16-bit color values, so each 8-bit component is
// multiplied by 257 (0xFF * 257 = 0xFFFF = 65535).
func HexToTerminalAppRGB(hex string) (string, error) {
	if err := ValidateHex(hex); err != nil {
		return "", err
	}

	hex = strings.TrimPrefix(hex, "#")
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)

	return fmt.Sprintf("{%d, %d, %d}", r*257, g*257, b*257), nil
}
