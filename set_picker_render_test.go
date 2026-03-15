package main

import (
	"strings"
	"testing"
)

func TestRenderPickerIncludesVisibleFilterAndViewMode(t *testing.T) {
	state := NewPickerState([]PickerItem{{Name: "dracula"}}, "dracula")
	state.SetQuery("dr")

	rendered := RenderPicker(*state, 100, 30)
	if !strings.Contains(rendered, "Filter: dr") {
		t.Fatalf("expected filter line, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "View: all") {
		t.Fatalf("expected view mode, got:\n%s", rendered)
	}
}

func TestRenderPickerIncludesFavoritesAndUsageBadge(t *testing.T) {
	state := NewPickerState([]PickerItem{{Name: "dracula", Favorite: true, UsageCount: 7}}, "dracula")

	rendered := RenderPicker(*state, 100, 30)
	if !strings.Contains(rendered, "* dracula") {
		t.Fatalf("expected favorite marker, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "used in 7 dirs") {
		t.Fatalf("expected usage badge, got:\n%s", rendered)
	}
}

func TestRenderPickerIncludesIntegratedPreviewSections(t *testing.T) {
	state := NewPickerState([]PickerItem{{Name: "dracula"}}, "dracula")

	rendered := RenderPicker(*state, 100, 30)
	for _, needle := range []string{"sample.zig", "less README.md", "NOTES.md", "previewing dracula"} {
		if !strings.Contains(rendered, needle) {
			t.Fatalf("expected %q in render:\n%s", needle, rendered)
		}
	}
}
