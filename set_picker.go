package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jkleemann/coltty/adapter"
	"github.com/spf13/cobra"
)

var interactiveSetRunner = runInteractiveSet

type pickerRuntime struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	isTTY   func() bool
	getwd   func() (string, error)
	homeDir func() (string, error)

	loadGlobalConfig func() (*GlobalConfig, error)
	findDirConfig    func(string) (string, *DirConfig, error)
	resolveCurrent   func(string, *GlobalConfig) (*ResolvedScheme, error)
	loadFavorites    func() (*FavoritesConfig, error)
	saveFavorites    func(*FavoritesConfig) error
	scanUsage        func(string) (map[string]int, error)

	applier      PreviewApplier
	startProgram func(pickerModel) (pickerModel, error)
}

type pickerEffects struct {
	onPreview       func(string) error
	onConfirm       func(string) error
	onCancel        func() error
	onSaveFavorites func([]string) error
}

func defaultPickerRuntime() pickerRuntime {
	return pickerRuntime{
		stdin:            os.Stdin,
		stdout:           os.Stdout,
		stderr:           os.Stderr,
		isTTY:            stdinIsTTY,
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

	model := newPickerModel(state, nil)
	model.effects = pickerEffects{
		onPreview: func(name string) error {
			item, ok := state.ItemByName(name)
			if !ok {
				return nil
			}
			return preview.ApplySelection(ResolvedFromScheme("", item.Name, item.Scheme))
		},
		onConfirm: func(name string) error {
			item, ok := state.ItemByName(name)
			if !ok {
				return fmt.Errorf("unknown picker selection %q", name)
			}
			configPath := filepath.Join(".", dirConfigFile)
			if err := WriteDirSchemeConfig(configPath, item.Name, item.Scheme, setInline); err != nil {
				return err
			}
			return preview.Confirm(ResolvedFromScheme(configPath, item.Name, item.Scheme))
		},
		onCancel: preview.Cancel,
		onSaveFavorites: func(names []string) error {
			return rt.saveFavorites(&FavoritesConfig{Schemes: names})
		},
	}

	if rt.startProgram != nil {
		_, err := rt.startProgram(model)
		return err
	}

	p := tea.NewProgram(
		model,
		tea.WithInput(rt.stdin),
		tea.WithOutput(rt.stdout),
		tea.WithAltScreen(),
	)
	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("starting picker: %w", err)
	}
	return nil
}

func stdinIsTTY() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
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

func (s *PickerState) ItemByName(name string) (PickerItem, bool) {
	for _, item := range s.Items {
		if item.Name == name {
			return item, true
		}
	}
	return PickerItem{}, false
}

func testProgramFromInput(input []byte) func(pickerModel) (pickerModel, error) {
	return func(model pickerModel) (pickerModel, error) {
		current := model
		cmds := []tea.Cmd{}

		reader := bufio.NewReader(bytes.NewReader(input))
		for {
			for len(cmds) > 0 {
				cmd := cmds[0]
				cmds = cmds[1:]
				if cmd == nil {
					continue
				}
				msg := cmd()
				if msg == nil {
					continue
				}
				if _, ok := msg.(tea.QuitMsg); ok {
					return current, nil
				}
				next, cmd := current.Update(msg)
				current = next.(pickerModel)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}

			msg, err := readTestTeaKey(reader)
			if err != nil {
				if err == io.EOF {
					return current, nil
				}
				return current, err
			}
			next, cmd := current.Update(msg)
			current = next.(pickerModel)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
}

func readTestTeaKey(reader *bufio.Reader) (tea.Msg, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case '\r', '\n':
		return tea.KeyMsg{Type: tea.KeyEnter}, nil
	case '\t':
		return tea.KeyMsg{Type: tea.KeyTab}, nil
	case 127, 8:
		return tea.KeyMsg{Type: tea.KeyBackspace}, nil
	case 27:
		if reader.Buffered() >= 2 {
			next, _ := reader.ReadByte()
			if next == '[' {
				dir, _ := reader.ReadByte()
				switch dir {
				case 'A':
					return tea.KeyMsg{Type: tea.KeyUp}, nil
				case 'B':
					return tea.KeyMsg{Type: tea.KeyDown}, nil
				}
			}
		}
		return tea.KeyMsg{Type: tea.KeyEsc}, nil
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(b)}}, nil
	}
}
