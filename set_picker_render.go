package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	if model.height == 0 {
		model.height = 24
	}

	leftWidth := maxInt(28, model.width/3)
	rightWidth := maxInt(40, model.width-leftWidth-3)
	contentHeight := model.contentHeight()
	leftContentWidth := maxInt(1, leftWidth-6)
	rightContentWidth := maxInt(1, rightWidth-6)

	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Height(contentHeight).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		Render(renderLeftPane(model, leftContentWidth))
	rightPane := lipgloss.NewStyle().
		Width(rightWidth).
		Height(contentHeight).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		Render(renderRightPane(model, rightContentWidth))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

func renderLeftPane(model pickerModel, contentWidth int) string {
	lines := []string{
		"Filter",
		model.input.View(),
		"",
		fmt.Sprintf("View: %s", model.state.ViewMode),
		"Themes",
	}

	if len(model.state.Filtered) == 0 {
		lines = append(lines, "no matches")
		if model.status != "" {
			lines = append(lines, "", model.status)
		}
		return fitLines(lines, model.contentHeight(), contentWidth)
	}

	rows := model.listViewportHeight()
	start := model.scrollOffset
	end := minInt(len(model.state.Filtered), start+rows)
	for i := start; i < end; i++ {
		index := model.state.Filtered[i]
		item := model.state.Items[index]
		prefix := "  "
		if i == model.state.Selected {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%s %s", prefix, favoriteMarker(item.Favorite), item.Name)
		if item.UsageCount > 0 {
			line += fmt.Sprintf("  used in %d dirs", item.UsageCount)
		}
		lines = append(lines, line)
	}

	if model.status != "" {
		lines = append(lines, "", model.status)
	}
	return fitLines(lines, model.contentHeight(), contentWidth)
}

func renderRightPane(model pickerModel, contentWidth int) string {
	selected := model.state.SelectedItem()
	name := selected.Name
	if name == "" {
		name = model.previewName
	}
	if name == "" {
		name = "none"
	}

	roles := newPreviewStyleRoles(selected.Scheme)
	lines := renderPreviewLines(model, roles, name)
	return fitLines(lines, model.contentHeight(), contentWidth)
}

func renderPreviewLines(model pickerModel, roles previewStyleRoles, name string) []string {
	lines := []string{
		roles.Heading.Render("Preview"),
		roles.Base.Render(model.selectedSchemeTitle()),
		"",
	}
	lines = append(lines, renderPalettePreview(model.state.SelectedItem(), roles)...)
	lines = append(lines, "")
	lines = append(lines, renderZigPreview(name, roles)...)
	lines = append(lines, "")
	lines = append(lines, renderLessPreview(roles)...)
	lines = append(lines, "")
	lines = append(lines, renderMarkdownPreview(roles)...)
	return lines
}

func renderPalettePreview(selected PickerItem, roles previewStyleRoles) []string {
	lines := []string{roles.Heading.Render("Palette")}
	if len(selected.Scheme.Palette) == 0 {
		return append(lines, roles.Muted.Render("(no palette)"))
	}

	chips := make([]string, 0, minInt(len(selected.Scheme.Palette), 8))
	for _, color := range selected.Scheme.Palette[:minInt(len(selected.Scheme.Palette), 8)] {
		chips = append(chips, lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(color))
	}
	return append(lines, strings.Join(chips, " "))
}

func renderZigPreview(name string, roles previewStyleRoles) []string {
	return []string{
		roles.Muted.Render("sample.zig"),
		roles.Keyword.Render("const") + " " + roles.Function.Render("std") + roles.Base.Render(" = @import(") + roles.String.Render("\"std\"") + roles.Base.Render(");"),
		roles.Keyword.Render("pub fn") + " " + roles.Function.Render("main") + roles.Base.Render("() !void {"),
		"    " + roles.Keyword.Render("const") + " " + roles.Function.Render("theme_name") + roles.Base.Render(" = ") + roles.String.Render(fmt.Sprintf("\"%s\"", name)),
		"    " + roles.Keyword.Render("const") + " " + roles.Function.Render("sample") + roles.Base.Render(" = .{ .ok = true, .depth = 3 };"),
		"    " + roles.Accent.Render("try") + roles.Base.Render(" std.debug.print(") + roles.String.Render("\"theme: {s} depth={d}\\n\"") + roles.Base.Render(", .{ theme_name, sample.depth });"),
		roles.Base.Render("}"),
	}
}

func renderLessPreview(roles previewStyleRoles) []string {
	return []string{
		roles.Muted.Render("less README.md"),
		roles.Heading.Render("NAME"),
		"  " + roles.Function.Render("coltty set") + roles.Base.Render(" - interactive theme selection"),
		roles.Heading.Render("USAGE"),
		"  " + roles.Accent.Render("$") + roles.Base.Render(" coltty set  ") + roles.Muted.Render("# arrows move, Enter saves"),
		roles.Heading.Render("STATUS"),
		"  " + roles.Bullet.Render("preview") + roles.Base.Render(": live, transient, restorable"),
	}
}

func renderMarkdownPreview(roles previewStyleRoles) []string {
	return []string{
		roles.Muted.Render("NOTES.md"),
		roles.Heading.Render("# Preview Behavior"),
		roles.Bullet.Render("-") + roles.Base.Render(" `Enter` saves the selected theme"),
		roles.Bullet.Render("-") + roles.Base.Render(" `Esc` restores the original colors"),
		roles.Bullet.Render("-") + roles.Base.Render(" `f` toggles favorites and `Tab` switches view"),
	}
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

func fitLines(lines []string, height int, width int) string {
	if height < 1 {
		height = 1
	}
	if width < 1 {
		width = 1
	}
	for i, line := range lines {
		lines[i] = truncateLine(line, width)
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func truncateLine(line string, width int) string {
	if ansi.StringWidth(line) <= width {
		return line
	}
	if width <= 1 {
		return ansi.Truncate(line, width, "")
	}
	return ansi.Truncate(line, width, "…")
}

func favoriteMarker(favorite bool) string {
	if favorite {
		return "*"
	}
	return " "
}
