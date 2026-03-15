package main

import (
	"fmt"
	"strings"
)

func RenderPicker(state PickerState, width, height int) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Filter: %s\n", state.Query)
	fmt.Fprintf(&b, "View: %s\n", state.ViewMode)
	b.WriteString("Themes:\n")

	for i, index := range state.Filtered {
		item := state.Items[index]
		prefix := "  "
		if i == state.Selected {
			prefix = "> "
		}
		favorite := " "
		if item.Favorite {
			favorite = "*"
		}
		fmt.Fprintf(&b, "%s%s %s", prefix, favorite, item.Name)
		if item.UsageCount > 0 {
			fmt.Fprintf(&b, "  used in %d dirs", item.UsageCount)
		}
		b.WriteString("\n")
	}

	selected := state.SelectedItem()
	b.WriteString("\nPreview:\n")
	fmt.Fprintf(&b, "previewing %s\n", selected.Name)
	b.WriteString("sample.zig\n")
	b.WriteString("const std = @import(\"std\");\n")
	b.WriteString("less README.md\n")
	b.WriteString("USAGE\n")
	b.WriteString("NOTES.md\n")
	b.WriteString("# Preview Behavior\n")

	return b.String()
}
