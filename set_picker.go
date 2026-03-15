package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/jkleemann/coltty/adapter"
	"github.com/spf13/cobra"
)

var interactiveSetRunner = runInteractiveSet

type pickerRuntime struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	isTTY   func() bool
	makeRaw func() (func(), error)
	getwd   func() (string, error)
	homeDir func() (string, error)

	loadGlobalConfig func() (*GlobalConfig, error)
	findDirConfig    func(string) (string, *DirConfig, error)
	resolveCurrent   func(string, *GlobalConfig) (*ResolvedScheme, error)
	loadFavorites    func() (*FavoritesConfig, error)
	saveFavorites    func(*FavoritesConfig) error
	scanUsage        func(string) (map[string]int, error)

	applier PreviewApplier
}

func defaultPickerRuntime() pickerRuntime {
	return pickerRuntime{
		stdin:            os.Stdin,
		stdout:           os.Stdout,
		stderr:           os.Stderr,
		isTTY:            stdinIsTTY,
		makeRaw:          makeRawTerminal,
		getwd:            os.Getwd,
		homeDir:          os.UserHomeDir,
		loadGlobalConfig: LoadGlobalConfig,
		findDirConfig:    FindDirConfig,
		resolveCurrent:   Resolve,
		loadFavorites:    LoadFavorites,
		saveFavorites:    SaveFavorites,
		scanUsage:        ScanThemeUsage,
		applier:          terminalPreviewApplier{},
	}
}

type terminalPreviewApplier struct{}

func (terminalPreviewApplier) Apply(scheme *ResolvedScheme) error {
	a := adapter.DetectAdapter(adapter.AllAdapters())
	if a == nil {
		return nil
	}
	return a.Apply(toAdapterScheme(scheme))
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
	a := adapter.DetectAdapter(adapter.AllAdapters())
	if a != nil {
		if err := a.Apply(toAdapterScheme(resolved)); err != nil {
			fmt.Fprintf(os.Stderr, "coltty: warning: %s adapter: %v\n", a.Name(), err)
		}
	}

	fmt.Fprintf(os.Stderr, "coltty: set scheme %q in %s\n", schemeName, dirConfigFile)
	return nil
}

func runInteractiveSet() error {
	return runInteractiveSetWithRuntime(defaultPickerRuntime())
}

func runInteractiveSetWithRuntime(rt pickerRuntime) error {
	if !rt.isTTY() {
		return fmt.Errorf("interactive picker requires a TTY; use 'coltty set <scheme>' instead")
	}

	restoreRaw, err := rt.makeRaw()
	if err != nil {
		return err
	}
	defer restoreRaw()

	fmt.Fprint(rt.stdout, "\x1b[?1049h\x1b[?25l")
	defer fmt.Fprint(rt.stdout, "\x1b[?25h\x1b[?1049l")

	globalCfg, err := rt.loadGlobalConfig()
	if err != nil {
		fmt.Fprintf(rt.stderr, "coltty: warning: failed to load global config: %v\n", err)
	}

	cwd, err := rt.getwd()
	if err != nil {
		return err
	}

	_, dirCfg, err := rt.findDirConfig(cwd)
	if err != nil {
		return err
	}

	original, err := rt.resolveCurrent(cwd, globalCfg)
	if err != nil {
		return err
	}

	favoritesCfg, err := rt.loadFavorites()
	if err != nil {
		fmt.Fprintf(rt.stderr, "coltty: warning: failed to load favorites: %v\n", err)
		favoritesCfg = &FavoritesConfig{}
	}
	favoriteSet := make(map[string]bool, len(favoritesCfg.Schemes))
	for _, scheme := range favoritesCfg.Schemes {
		favoriteSet[scheme] = true
	}

	usage := map[string]int{}
	if home, err := rt.homeDir(); err == nil {
		if counts, err := rt.scanUsage(home); err == nil {
			usage = counts
		}
	}

	available := AvailableSchemes(globalCfg)
	items := make([]PickerItem, 0, len(available))
	for _, scheme := range available {
		items = append(items, PickerItem{
			Name:       scheme.Name,
			Scheme:     scheme.Scheme,
			Tag:        scheme.Tag,
			Favorite:   favoriteSet[scheme.Name],
			UsageCount: usage[scheme.Name],
		})
	}

	initialName := original.SchemeName
	if dirCfg != nil && dirCfg.Scheme == "" {
		if inferred, ok := InferClosestScheme(dirCfg.Overrides, globalCfg); ok {
			initialName = inferred
		}
	}
	if initialName == "" && len(items) > 0 {
		initialName = items[0].Name
	}

	state := NewPickerState(items, initialName)
	preview := NewPreviewSession(rt.applier, original)
	if item := state.SelectedItem(); item.Name != "" {
		if err := preview.ApplySelection(ResolvedFromScheme("", item.Name, item.Scheme)); err != nil {
			fmt.Fprintf(rt.stderr, "coltty: warning: preview apply failed: %v\n", err)
		}
	}

	reader := bufio.NewReader(rt.stdin)
	for {
		fmt.Fprint(rt.stdout, "\x1b[H\x1b[2J")
		fmt.Fprint(rt.stdout, RenderPicker(*state, 0, 0))

		key, err := readPickerKey(reader)
		if err != nil {
			if err == io.EOF {
				return preview.Cancel()
			}
			return err
		}

		switch key {
		case keyArrowUp:
			if state.MoveSelection(-1) {
				_ = preview.ApplySelection(ResolvedFromScheme("", state.SelectedItem().Name, state.SelectedItem().Scheme))
			}
		case keyArrowDown:
			if state.MoveSelection(1) {
				_ = preview.ApplySelection(ResolvedFromScheme("", state.SelectedItem().Name, state.SelectedItem().Scheme))
			}
		case keyEnter:
			item := state.SelectedItem()
			configPath := filepath.Join(".", dirConfigFile)
			if err := WriteDirSchemeConfig(configPath, item.Name, item.Scheme, setInline); err != nil {
				return err
			}
			if err := preview.Confirm(ResolvedFromScheme(configPath, item.Name, item.Scheme)); err != nil {
				return err
			}
			return nil
		case keyEscape:
			if state.Query != "" {
				state.SetQuery("")
				continue
			}
			return preview.Cancel()
		case keyBackspace:
			if state.Query != "" {
				state.SetQuery(state.Query[:len(state.Query)-1])
			}
		case keyTab:
			state.ToggleViewMode()
		case keyFavorite:
			state.ToggleFavorite()
			if err := rt.saveFavorites(&FavoritesConfig{Schemes: state.FavoriteNames()}); err != nil {
				fmt.Fprintf(rt.stderr, "coltty: warning: failed to save favorites: %v\n", err)
			}
		default:
			if keyPrintable(key) {
				state.SetQuery(state.Query + key)
			}
		}
	}
}

const (
	keyArrowUp   = "up"
	keyArrowDown = "down"
	keyEnter     = "enter"
	keyEscape    = "escape"
	keyBackspace = "backspace"
	keyTab       = "tab"
	keyFavorite  = "favorite"
)

func readPickerKey(reader *bufio.Reader) (string, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return "", err
	}

	switch b {
	case '\r', '\n':
		return keyEnter, nil
	case '\t':
		return keyTab, nil
	case 127, 8:
		return keyBackspace, nil
	case 'f':
		return keyFavorite, nil
	case 27:
		if reader.Buffered() >= 2 {
			next, _ := reader.ReadByte()
			if next == '[' {
				dir, _ := reader.ReadByte()
				switch dir {
				case 'A':
					return keyArrowUp, nil
				case 'B':
					return keyArrowDown, nil
				}
			}
			return keyEscape, nil
		}
		return keyEscape, nil
	default:
		return string(b), nil
	}
}

func keyPrintable(key string) bool {
	return len(key) == 1 && key[0] >= 32 && key[0] < 127
}

func stdinIsTTY() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func makeRawTerminal() (func(), error) {
	get := exec.Command("stty", "-g")
	get.Stdin = os.Stdin
	state, err := get.Output()
	if err != nil {
		return nil, fmt.Errorf("reading terminal state: %w", err)
	}

	raw := exec.Command("stty", "raw", "-echo")
	raw.Stdin = os.Stdin
	if err := raw.Run(); err != nil {
		return nil, fmt.Errorf("enabling raw mode: %w", err)
	}

	saved := string(state)
	restore := func() {
		restoreCmd := exec.Command("stty", saved)
		restoreCmd.Stdin = os.Stdin
		_ = restoreCmd.Run()
	}
	return restore, nil
}

func (s *PickerState) MoveSelection(delta int) bool {
	if len(s.Filtered) == 0 {
		return false
	}
	next := s.Selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(s.Filtered) {
		next = len(s.Filtered) - 1
	}
	if next == s.Selected {
		return false
	}
	s.Selected = next
	return true
}

func (s *PickerState) FavoriteNames() []string {
	var favorites []string
	for _, item := range s.Items {
		if item.Favorite {
			favorites = append(favorites, item.Name)
		}
	}
	slices.Sort(favorites)
	return favorites
}
