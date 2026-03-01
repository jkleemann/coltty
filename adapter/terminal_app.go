package adapter

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// TerminalAppAdapter applies color schemes to macOS Terminal.app by creating
// or updating a named settings profile and switching to it via AppleScript.
type TerminalAppAdapter struct {
	// RunAppleScript executes an AppleScript string. Injectable for testing.
	// If nil, defaults to running osascript.
	RunAppleScript func(script string) error
}

// NewTerminalAppAdapter creates a TerminalAppAdapter with the default osascript runner.
func NewTerminalAppAdapter() *TerminalAppAdapter {
	return &TerminalAppAdapter{}
}

func (a *TerminalAppAdapter) Name() string {
	return "terminal.app"
}

func (a *TerminalAppAdapter) Detect() bool {
	return os.Getenv("TERM_PROGRAM") == "Apple_Terminal"
}

func (a *TerminalAppAdapter) Apply(scheme *ResolvedScheme) error {
	profile := scheme.Name
	if v, ok := scheme.Extras["terminal_app_profile"]; ok && v != "" {
		profile = v
	}

	if profile == "" {
		return fmt.Errorf("terminal.app adapter requires a scheme name or terminal_app_profile")
	}

	script, err := BuildTerminalAppApplyScript(profile, scheme)
	if err != nil {
		return err
	}

	runner := a.RunAppleScript
	if runner == nil {
		runner = DefaultRunAppleScript
	}
	return runner(script)
}

// DefaultRunAppleScript executes an AppleScript string via osascript.
// The script is passed via stdin to handle multi-line scripts correctly.
func DefaultRunAppleScript(script string) error {
	cmd := exec.Command("osascript")
	cmd.Stdin = strings.NewReader(script)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// BuildTerminalAppApplyScript generates AppleScript that creates/updates a
// Terminal.app profile with the given colors and switches the front window to it.
func BuildTerminalAppApplyScript(profileName string, scheme *ResolvedScheme) (string, error) {
	return buildScript(profileName, scheme, true)
}

// BuildTerminalAppSetupScript generates AppleScript that creates/updates a
// Terminal.app profile with the given colors, without switching to it.
func BuildTerminalAppSetupScript(profileName string, scheme *ResolvedScheme) (string, error) {
	return buildScript(profileName, scheme, false)
}

func buildScript(profileName string, scheme *ResolvedScheme, switchProfile bool) (string, error) {
	var b strings.Builder

	b.WriteString("tell application \"Terminal\"\n")

	// Ensure the profile exists by duplicating the first settings set if needed.
	fmt.Fprintf(&b, "\tset profileNames to name of every settings set\n")
	fmt.Fprintf(&b, "\tif %q is not in profileNames then\n", profileName)
	fmt.Fprintf(&b, "\t\tmake new settings set with properties {name:%q}\n", profileName)
	fmt.Fprintf(&b, "\tend if\n")
	fmt.Fprintf(&b, "\tset targetProfile to settings set %q\n", profileName)

	if err := writeColorProperties(&b, scheme); err != nil {
		return "", err
	}

	if switchProfile {
		b.WriteString("\tset current settings of front window to targetProfile\n")
	}

	b.WriteString("end tell\n")

	return b.String(), nil
}

func writeColorProperties(b *strings.Builder, scheme *ResolvedScheme) error {
	type colorProp struct {
		property string
		hex      string
	}

	props := []colorProp{
		{"normal text color", scheme.Foreground},
		{"background color", scheme.Background},
		{"cursor color", scheme.Cursor},
	}

	for _, p := range props {
		if p.hex == "" {
			continue
		}
		rgb, err := HexToTerminalAppRGB(p.hex)
		if err != nil {
			return fmt.Errorf("converting %s %q: %w", p.property, p.hex, err)
		}
		fmt.Fprintf(b, "\tset %s of targetProfile to %s\n", p.property, rgb)
	}

	return nil
}
