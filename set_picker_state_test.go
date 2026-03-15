package main

import "testing"

func TestPickerStateInitialSelectionUsesActiveScheme(t *testing.T) {
	state := NewPickerState([]PickerItem{
		{Name: "gruvbox"},
		{Name: "dracula"},
		{Name: "nord"},
	}, "dracula")

	if got := state.SelectedItem().Name; got != "dracula" {
		t.Fatalf("expected dracula to be selected, got %q", got)
	}
}

func TestPickerStateFilterUsesFuzzyMatching(t *testing.T) {
	state := NewPickerState([]PickerItem{
		{Name: "dracula"},
		{Name: "solarized-dark"},
		{Name: "nord"},
	}, "")

	state.SetQuery("sd")

	if len(state.Filtered) != 1 {
		t.Fatalf("expected one fuzzy match, got %d", len(state.Filtered))
	}
	if got := state.SelectedItem().Name; got != "solarized-dark" {
		t.Fatalf("expected solarized-dark, got %q", got)
	}
}

func TestPickerStateFilterMovesSelectionToTopMatch(t *testing.T) {
	state := NewPickerState([]PickerItem{
		{Name: "dracula"},
		{Name: "solarized-dark"},
		{Name: "rose-pine"},
	}, "rose-pine")

	state.SetQuery("dr")

	if got := state.SelectedItem().Name; got != "dracula" {
		t.Fatalf("expected selection to move to dracula, got %q", got)
	}
}

func TestPickerStateToggleFavorite(t *testing.T) {
	state := NewPickerState([]PickerItem{
		{Name: "dracula"},
	}, "dracula")

	state.ToggleFavorite()

	if !state.Items[0].Favorite {
		t.Fatal("expected item to be favorite")
	}
}

func TestPickerStateToggleViewFavoritesOnly(t *testing.T) {
	state := NewPickerState([]PickerItem{
		{Name: "dracula", Favorite: true},
		{Name: "nord", Favorite: false},
	}, "dracula")

	state.ToggleViewMode()

	if state.ViewMode != ViewFavorites {
		t.Fatalf("expected favorites view, got %q", state.ViewMode)
	}
	if len(state.Filtered) != 1 {
		t.Fatalf("expected one favorite item, got %d", len(state.Filtered))
	}
	if got := state.SelectedItem().Name; got != "dracula" {
		t.Fatalf("expected dracula in favorites view, got %q", got)
	}
}
