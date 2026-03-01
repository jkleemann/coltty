package adapter

import (
	"fmt"
	"strconv"
	"strings"
)

// HexToTerminalAppRGB converts a hex color string like "#f8f8f2" to a
// Terminal.app AppleScript RGB list like "{63736, 63736, 62194}".
// Terminal.app uses 16-bit color values, so each 8-bit component is
// multiplied by 257 (0xFF * 257 = 0xFFFF = 65535).
func HexToTerminalAppRGB(hex string) (string, error) {
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return "", fmt.Errorf("invalid hex color %q: expected 6 hex digits", hex)
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return "", fmt.Errorf("invalid hex color %q: %w", hex, err)
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return "", fmt.Errorf("invalid hex color %q: %w", hex, err)
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return "", fmt.Errorf("invalid hex color %q: %w", hex, err)
	}

	return fmt.Sprintf("{%d, %d, %d}", r*257, g*257, b*257), nil
}
