package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPickerModelWindowSizeLaysOutTwoPanes(t *testing.T) {
	model := newPickerModel(newPickerStateFixture(), nil)
	next, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model = next.(pickerModel)

	view := model.View()
	for _, needle := range []string{"Filter", "Themes", "Preview", "sample.zig"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected %q in view:\n%s", needle, view)
		}
	}
}

func TestPickerModelTypingUpdatesVisibleFilter(t *testing.T) {
	model := newPickerModel(newPickerStateFixture(), nil)
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	model = next.(pickerModel)
	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model = next.(pickerModel)

	if got := model.input.Value(); got != "dr" {
		t.Fatalf("expected filter dr, got %q", got)
	}
	if got := model.state.SelectedItem().Name; got != "dracula" {
		t.Fatalf("expected dracula selected, got %q", got)
	}
}

func TestPickerModelTabTogglesFavoritesView(t *testing.T) {
	state := NewPickerState([]PickerItem{
		{Name: "dracula", Favorite: true},
		{Name: "nord"},
	}, "dracula")
	model := newPickerModel(state, nil)

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = next.(pickerModel)

	if model.state.ViewMode != ViewFavorites {
		t.Fatalf("expected favorites view, got %q", model.state.ViewMode)
	}
	if len(model.state.Filtered) != 1 {
		t.Fatalf("expected one filtered favorite, got %d", len(model.state.Filtered))
	}
}

func TestPickerModelSelectionChangeEmitsPreviewIntent(t *testing.T) {
	model := newPickerModel(newPickerStateFixture(), nil)
	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = next.(pickerModel)

	msg := cmd()
	preview, ok := msg.(previewSelectionMsg)
	if !ok {
		t.Fatalf("expected previewSelectionMsg, got %#v", msg)
	}
	if preview.schemeName == "" {
		t.Fatal("expected preview scheme name")
	}
	if model.state.SelectedItem().Name != preview.schemeName {
		t.Fatalf("expected preview for %q, got %q", model.state.SelectedItem().Name, preview.schemeName)
	}
}

func TestPickerModelEnterEmitsConfirmIntent(t *testing.T) {
	model := newPickerModel(newPickerStateFixture(), nil)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg := cmd()

	confirm, ok := msg.(confirmSelectionMsg)
	if !ok {
		t.Fatalf("expected confirmSelectionMsg, got %#v", msg)
	}
	if confirm.schemeName != model.state.SelectedItem().Name {
		t.Fatalf("expected confirm for %q, got %q", model.state.SelectedItem().Name, confirm.schemeName)
	}
}

func TestPickerModelEscapeClearsFilterBeforeCancel(t *testing.T) {
	model := newPickerModel(newPickerStateFixture(), nil)
	model.input.SetValue("dr")
	model.state.SetQuery("dr")

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = next.(pickerModel)

	if model.input.Value() != "" {
		t.Fatalf("expected filter cleared, got %q", model.input.Value())
	}
	if model.state.Query != "" {
		t.Fatalf("expected state query cleared, got %q", model.state.Query)
	}
	if cmd != nil {
		if msg := cmd(); msg != nil {
			t.Fatalf("expected no cancel message on first escape, got %#v", msg)
		}
	}
}

func TestPickerModelNoMatchesShowsEmptyStateWithoutDroppingPreview(t *testing.T) {
	model := newPickerModel(newPickerStateFixture(), nil)
	model.previewName = "dracula"
	model.input.SetValue("zzz")
	model.state.SetQuery("zzz")

	view := model.View()
	if !strings.Contains(view, "no matches") {
		t.Fatalf("expected no matches state, got:\n%s", view)
	}
	if !strings.Contains(view, "previewing dracula") {
		t.Fatalf("expected previous preview to remain visible, got:\n%s", view)
	}
}

func TestPickerModelSelectionScrollsListViewport(t *testing.T) {
	items := []PickerItem{
		{Name: "alpha", Scheme: BuiltinSchemes()["dracula"]},
		{Name: "beta", Scheme: BuiltinSchemes()["dracula"]},
		{Name: "gamma", Scheme: BuiltinSchemes()["dracula"]},
		{Name: "delta", Scheme: BuiltinSchemes()["dracula"]},
		{Name: "epsilon", Scheme: BuiltinSchemes()["dracula"]},
		{Name: "zeta", Scheme: BuiltinSchemes()["dracula"]},
		{Name: "eta", Scheme: BuiltinSchemes()["dracula"]},
	}
	model := newPickerModel(NewPickerState(items, "alpha"), nil)
	next, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 12})
	model = next.(pickerModel)

	for i := 0; i < 5; i++ {
		next, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = next.(pickerModel)
	}

	view := model.View()
	if !strings.Contains(view, ">   zeta") && !strings.Contains(view, "> * zeta") && !strings.Contains(view, ">   eta") {
		t.Fatalf("expected lower selection visible in viewport, got:\n%s", view)
	}
	if strings.Contains(view, "alpha") && strings.Contains(view, "zeta") {
		t.Fatalf("expected top of long list to scroll out of view, got:\n%s", view)
	}
}

func TestRenderPickerViewRespectsTerminalHeight(t *testing.T) {
	model := newPickerModel(newPickerStateFixture(), nil)
	model.width = 100
	model.height = 14

	rendered := renderPickerView(model)
	lines := strings.Split(rendered, "\n")
	if len(lines) > 14 {
		t.Fatalf("expected render to fit terminal height, got %d lines:\n%s", len(lines), rendered)
	}
}

func newPickerStateFixture() *PickerState {
	return NewPickerState([]PickerItem{
		{Name: "catppuccin"},
		{Name: "dracula", Favorite: true, UsageCount: 3, Scheme: BuiltinSchemes()["dracula"]},
		{Name: "nord", UsageCount: 1, Scheme: BuiltinSchemes()["nord"]},
	}, "dracula")
}
