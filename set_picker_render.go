package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderPicker(state PickerState, width, height int) string {
	model := newPickerModel(&state, nil)
	model.width = width
	model.height = height
	return renderPickerView(model)
}

func renderPickerView(model pickerModel) string {
	if model.width == 0 {
		model.width = 100
	}

	leftWidth := maxInt(28, model.width/3)
	rightWidth := maxInt(40, model.width-leftWidth-3)

	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		Render(renderLeftPane(model))
	rightPane := lipgloss.NewStyle().
		Width(rightWidth).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		Render(renderRightPane(model))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

func renderLeftPane(model pickerModel) string {
	var b strings.Builder

	b.WriteString("Filter\n")
	b.WriteString(model.input.View())
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "View: %s\n", model.state.ViewMode)
	b.WriteString("Themes\n")

	if len(model.state.Filtered) == 0 {
		b.WriteString("no matches")
		return b.String()
	}

	for i, index := range model.state.Filtered {
		item := model.state.Items[index]
		prefix := "  "
		if i == model.state.Selected {
			prefix = "> "
		}
		favorite := " "
		if item.Favorite {
			favorite = "*"
		}
		line := fmt.Sprintf("%s%s %s", prefix, favorite, item.Name)
		if item.UsageCount > 0 {
			line += fmt.Sprintf("  used in %d dirs", item.UsageCount)
		}
		b.WriteString(line + "\n")
	}

	if model.status != "" {
		b.WriteString("\n" + model.status)
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderRightPane(model pickerModel) string {
	selected := model.state.SelectedItem()
	name := selected.Name
	if name == "" {
		name = model.previewName
	}
	if name == "" {
		name = "none"
	}

	palette := []string{}
	if len(selected.Scheme.Palette) > 0 {
		palette = selected.Scheme.Palette[:minInt(len(selected.Scheme.Palette), 8)]
	}

	var b strings.Builder
	b.WriteString("Preview\n")
	b.WriteString(model.selectedSchemeTitle())
	b.WriteString("\n\nPalette\n")
	if len(palette) == 0 {
		b.WriteString("(no palette)\n")
	} else {
		b.WriteString(strings.Join(palette, " "))
		b.WriteString("\n")
	}
	b.WriteString("\nsample.zig\n")
	b.WriteString("const std = @import(\"std\");\n")
	b.WriteString("pub fn main() !void {\n")
	b.WriteString("    try std.debug.print(\"theme: {s}\\n\", .{\"" + name + "\"});\n")
	b.WriteString("}\n")
	b.WriteString("\nless README.md\n")
	b.WriteString("USAGE\n")
	b.WriteString("  coltty set\n")
	b.WriteString("\nNOTES.md\n")
	b.WriteString("# Preview Behavior\n")
	b.WriteString("- Enter saves\n")
	b.WriteString("- Esc restores\n")

	return strings.TrimRight(b.String(), "\n")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
