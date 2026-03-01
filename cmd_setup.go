package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/jkleemann/coltty/adapter"
	"github.com/spf13/cobra"
)

// setupRunAppleScript is injectable for testing. If nil, uses adapter.DefaultRunAppleScript.
var setupRunAppleScript func(script string) error

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "One-time setup commands for specific terminals",
}

var setupTerminalAppCmd = &cobra.Command{
	Use:   "terminal-app",
	Short: "Create Terminal.app profiles for all known color schemes",
	Long: `Creates or updates a Terminal.app settings profile for every color scheme
(built-in and user-defined). Each profile gets the scheme's foreground,
background, and cursor colors. Running this command again is idempotent —
existing profiles are updated in place.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		globalCfg, err := LoadGlobalConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "coltty: warning: failed to load global config: %v\n", err)
		}

		// Collect all schemes: built-in first, user overrides on top.
		schemes := make(map[string]Scheme)
		for name, s := range BuiltinSchemes() {
			schemes[name] = s
		}
		if globalCfg != nil {
			for name, s := range globalCfg.Schemes {
				schemes[name] = s
			}
		}

		if len(schemes) == 0 {
			fmt.Fprintln(os.Stderr, "coltty: no schemes found")
			return nil
		}

		// Sort names for deterministic output.
		names := make([]string, 0, len(schemes))
		for name := range schemes {
			names = append(names, name)
		}
		sort.Strings(names)

		runner := setupRunAppleScript
		if runner == nil {
			runner = adapter.DefaultRunAppleScript
		}

		var created int
		for _, name := range names {
			s := schemes[name]
			adapterScheme := &adapter.ResolvedScheme{
				Foreground: s.Foreground,
				Background: s.Background,
				Cursor:     s.Cursor,
				Name:       name,
			}

			script, err := adapter.BuildTerminalAppSetupScript(name, adapterScheme)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\u2717 %s: %v\n", name, err)
				continue
			}

			if err := runner(script); err != nil {
				fmt.Fprintf(os.Stderr, "\u2717 %s: %v\n", name, err)
				continue
			}

			fmt.Fprintf(os.Stderr, "\u2713 %s\n", name)
			created++
		}

		fmt.Fprintf(os.Stderr, "coltty: created/updated %d Terminal.app profiles\n", created)
		return nil
	},
}

func init() {
	setupCmd.AddCommand(setupTerminalAppCmd)
	rootCmd.AddCommand(setupCmd)
}
