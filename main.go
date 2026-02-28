package main

import (
	"fmt"
	"os"
	"sort"
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
	Short: "List all named schemes from the global config",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalCfg, err := LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("loading global config: %w", err)
		}

		if globalCfg == nil || len(globalCfg.Schemes) == 0 {
			fmt.Println("No schemes defined. Create a config at", globalConfigPath())
			return nil
		}

		defaultScheme := globalCfg.Default.Scheme

		names := make([]string, 0, len(globalCfg.Schemes))
		for name := range globalCfg.Schemes {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			s := globalCfg.Schemes[name]
			marker := ""
			if name == defaultScheme {
				marker = " (default)"
			}
			fmt.Printf("%s%s\n  fg: %s  bg: %s  cursor: %s\n", name, marker, s.Foreground, s.Background, s.Cursor)
		}

		return nil
	},
}

func init() {
	applyCmd.Flags().BoolVar(&applyQuiet, "quiet", false, "suppress output unless there's an error")
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "print the resolved scheme without applying")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(schemesCmd)
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
	return &adapter.ResolvedScheme{
		Foreground: s.Foreground,
		Background: s.Background,
		Cursor:     s.Cursor,
		Palette:    s.Palette,
	}
}
