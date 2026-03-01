package adapter

import (
	"fmt"
	"os"
	"os/exec"
)

// TerminalAppAdapter applies color schemes to macOS Terminal.app by switching
// to a named settings profile via AppleScript. OSC color-setting sequences
// don't work in Terminal.app, so profile switching is the only mechanism.
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

	script := fmt.Sprintf(
		`tell application "Terminal" to set current settings of front window to settings set "%s"`,
		profile,
	)

	runner := a.RunAppleScript
	if runner == nil {
		runner = defaultRunAppleScript
	}
	return runner(script)
}

func defaultRunAppleScript(script string) error {
	cmd := exec.Command("osascript", "-e", script)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
