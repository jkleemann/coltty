package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jkleemann/coltty/adapter"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "coltty",
	Short: "Automatically switch terminal color schemes based on directory",
	Long:  "Coltty is a CLI tool and shell hook that automatically switches terminal color schemes based on the current directory.",
}

var initCmd = &cobra.Command{
	Use:   "init <shell>",
	Short: "Print shell hook code for the given shell",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hook, err := ShellHook(args[0])
		if err != nil {
			return err
		}
		fmt.Print(hook)
		return nil
	},
}

var applyQuiet bool
var applyDryRun bool
var setInline bool

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the color scheme for the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalCfg, err := LoadGlobalConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "coltty: warning: failed to load global config: %v\n", err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}

		resolved, err := Resolve(cwd, globalCfg)
		if err != nil {
			return err
		}

		if applyDryRun {
			printScheme(resolved)
			return nil
		}

		adapterScheme := toAdapterScheme(resolved)

		if adapter.InScreen() && !applyQuiet {
			fmt.Fprintln(os.Stderr, "coltty: warning: GNU Screen does not support dynamic color changes")
		}

		a := adapter.DetectAdapter(adapter.AllAdapters())
		if a == nil {
			if !applyQuiet {
				fmt.Fprintln(os.Stderr, "coltty: no supported terminal detected")
			}
			return nil
		}

		if err := a.Apply(adapterScheme); err != nil {
			if !applyQuiet {
				fmt.Fprintf(os.Stderr, "coltty: warning: %s adapter: %v\n", a.Name(), err)
			}
			return nil
		}

		if !applyQuiet {
			fmt.Fprintf(os.Stderr, "coltty: applied scheme via %s (source: %s)\n", a.Name(), resolved.Source)
		}

		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the resolved color scheme for the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalCfg, err := LoadGlobalConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "coltty: warning: failed to load global config: %v\n", err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}

		resolved, err := Resolve(cwd, globalCfg)
		if err != nil {
			return err
		}

		printScheme(resolved)
		return nil
	},
}

var schemesCmd = &cobra.Command{
	Use:   "schemes",
	Short: "List all available schemes (built-in and user-defined)",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalCfg, err := LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("loading global config: %w", err)
		}

		var defaultScheme string
		if globalCfg != nil {
			defaultScheme = globalCfg.Default.Scheme
		}

		schemes := AvailableSchemes(globalCfg)
		if len(schemes) == 0 {
			fmt.Println("No schemes available.")
			return nil
		}

		for _, scheme := range schemes {
			marker := scheme.Tag
			if scheme.Name == defaultScheme {
				marker += " (default)"
			}
			fmt.Printf("%s%s\n  fg: %s  bg: %s  cursor: %s\n", scheme.Name, marker, scheme.Scheme.Foreground, scheme.Scheme.Background, scheme.Scheme.Cursor)
		}

		return nil
	},
}

var setCmd = &cobra.Command{
	Use:   "set [scheme]",
	Short: "Set the color scheme for the current directory",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetCommand(args)
	},
}

// formatInlineConfig generates a .coltty.toml with full color values under [overrides].
func formatInlineConfig(schemeName string, s Scheme) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Generated from scheme %q\n\n[overrides]\n", schemeName)
	fmt.Fprintf(&b, "foreground = %q\n", s.Foreground)
	fmt.Fprintf(&b, "background = %q\n", s.Background)
	fmt.Fprintf(&b, "cursor = %q\n", s.Cursor)
	if len(s.Palette) > 0 {
		b.WriteString("palette = [\n")
		for i := 0; i < len(s.Palette); i += 4 {
			end := i + 4
			if end > len(s.Palette) {
				end = len(s.Palette)
			}
			quoted := make([]string, end-i)
			for j, c := range s.Palette[i:end] {
				quoted[j] = fmt.Sprintf("%q", c)
			}
			b.WriteString("    ")
			b.WriteString(strings.Join(quoted, ", "))
			if end < len(s.Palette) {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString("]\n")
	}
	return b.String()
}

func init() {
	applyCmd.Flags().BoolVar(&applyQuiet, "quiet", false, "suppress output unless there's an error")
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "print the resolved scheme without applying")

	setCmd.Flags().BoolVar(&setInline, "inline", false, "write full color values instead of a scheme reference")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(schemesCmd)
	rootCmd.AddCommand(setCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printScheme(s *ResolvedScheme) {
	fmt.Printf("Source:     %s\n", s.Source)
	fmt.Printf("Foreground: %s\n", s.Foreground)
	fmt.Printf("Background: %s\n", s.Background)
	fmt.Printf("Cursor:     %s\n", s.Cursor)
	if len(s.Palette) > 0 {
		fmt.Printf("Palette:    %s\n", strings.Join(s.Palette, ", "))
	}
}

func toAdapterScheme(s *ResolvedScheme) *adapter.ResolvedScheme {
	extras := make(map[string]string)
	if s.Bold != "" {
		extras["bold"] = s.Bold
	}
	if s.SelectionForeground != "" {
		extras["selection_foreground"] = s.SelectionForeground
	}
	if s.SelectionBackground != "" {
		extras["selection_background"] = s.SelectionBackground
	}
	if s.Tab != "" {
		extras["tab"] = s.Tab
	}
	if s.ItermPreset != "" {
		extras["iterm_preset"] = s.ItermPreset
	}
	if s.TerminalAppProfile != "" {
		extras["terminal_app_profile"] = s.TerminalAppProfile
	}

	// Only set Extras if there are any values
	var extrasPtr map[string]string
	if len(extras) > 0 {
		extrasPtr = extras
	}

	return &adapter.ResolvedScheme{
		Foreground: s.Foreground,
		Background: s.Background,
		Cursor:     s.Cursor,
		Palette:    s.Palette,
		Name:       s.SchemeName,
		Extras:     extrasPtr,
	}
}
