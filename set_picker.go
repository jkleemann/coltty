package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jkleemann/coltty/adapter"
	"github.com/spf13/cobra"
)

var interactiveSetRunner = func() error {
	return fmt.Errorf("interactive picker not yet implemented")
}

func runSetCommand(args []string) error {
	switch len(args) {
	case 0:
		return interactiveSetRunner()
	case 1:
		return runDirectSet(args[0])
	default:
		return cobra.MaximumNArgs(1)(nil, args)
	}
}

func runDirectSet(schemeName string) error {
	globalCfg, err := LoadGlobalConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coltty: warning: failed to load global config: %v\n", err)
	}

	scheme, ok := LookupScheme(schemeName, globalCfg)
	if !ok {
		return fmt.Errorf("unknown scheme %q (use 'coltty schemes' to list available schemes)", schemeName)
	}

	configPath := filepath.Join(".", dirConfigFile)
	if _, err := os.Stat(configPath); err == nil {
		fmt.Fprintf(os.Stderr, "coltty: overwriting existing %s\n", dirConfigFile)
	}

	if err := WriteDirSchemeConfig(configPath, schemeName, scheme, setInline); err != nil {
		return fmt.Errorf("writing %s: %w", dirConfigFile, err)
	}

	resolved := ResolvedFromScheme(configPath, schemeName, scheme)
	adapterScheme := toAdapterScheme(resolved)

	a := adapter.DetectAdapter(adapter.AllAdapters())
	if a != nil {
		if err := a.Apply(adapterScheme); err != nil {
			fmt.Fprintf(os.Stderr, "coltty: warning: %s adapter: %v\n", a.Name(), err)
		}
	}

	fmt.Fprintf(os.Stderr, "coltty: set scheme %q in %s\n", schemeName, dirConfigFile)
	return nil
}
