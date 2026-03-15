package main

import (
	"strings"
	"testing"
)

func TestRenderPickerViewShowsFilterPaneAndPreviewPane(t *testing.T) {
	model := newPickerModel(newPickerStateFixture(), nil)
	model.width = 120
	model.height = 40

	rendered := renderPickerView(model)
	for _, needle := range []string{"Filter", "Themes", "Preview", "sample.zig"} {
		if !strings.Contains(rendered, needle) {
			t.Fatalf("expected %q in render:\n%s", needle, rendered)
		}
	}
}

func TestRenderPickerViewShowsFavoriteMarkerAndUsageBadge(t *testing.T) {
	model := newPickerModel(NewPickerState([]PickerItem{
		{Name: "dracula", Favorite: true, UsageCount: 7, Scheme: BuiltinSchemes()["dracula"]},
	}, "dracula"), nil)
	model.width = 120

	rendered := renderPickerView(model)
	if !strings.Contains(rendered, "* dracula") {
		t.Fatalf("expected favorite marker, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "used in 7 dirs") {
		t.Fatalf("expected usage badge, got:\n%s", rendered)
	}
}

func TestRenderPickerViewShowsIntegratedPreviewSections(t *testing.T) {
	model := newPickerModel(newPickerStateFixture(), nil)
	model.width = 120

	rendered := renderPickerView(model)
	for _, needle := range []string{"Palette", "sample.zig", "less README.md", "NOTES.md"} {
		if !strings.Contains(rendered, needle) {
			t.Fatalf("expected %q in render:\n%s", needle, rendered)
		}
	}
}
