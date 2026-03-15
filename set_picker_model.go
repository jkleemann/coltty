package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type pickerModel struct {
	width       int
	height      int
	input       textinput.Model
	state       *PickerState
	status      string
	effects     pickerEffects
	previewName string
}

type previewSelectionMsg struct {
	schemeName string
}

type confirmSelectionMsg struct {
	schemeName string
}

type cancelSelectionMsg struct{}

func newPickerModel(state *PickerState, err error) pickerModel {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "type to filter themes"
	input.Focus()

	model := pickerModel{
		input: input,
		state: state,
	}
	if selected := state.SelectedItem().Name; selected != "" {
		model.previewName = selected
	}
	if err != nil {
		model.status = err.Error()
	}
	return model
}

func (m pickerModel) Init() tea.Cmd {
	if name := m.state.SelectedItem().Name; name != "" {
		return tea.Batch(textinput.Blink, emitPreviewSelection(name))
	}
	return textinput.Blink
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case previewSelectionMsg:
		if msg.schemeName != "" {
			m.previewName = msg.schemeName
		}
		if m.effects.onPreview != nil {
			if err := m.effects.onPreview(msg.schemeName); err != nil {
				m.status = err.Error()
			}
		}
		return m, nil
	case confirmSelectionMsg:
		if m.effects.onConfirm != nil {
			if err := m.effects.onConfirm(msg.schemeName); err != nil {
				m.status = err.Error()
				return m, nil
			}
		}
		return m, tea.Quit
	case cancelSelectionMsg:
		if m.effects.onCancel != nil {
			if err := m.effects.onCancel(); err != nil {
				m.status = err.Error()
				return m, nil
			}
		}
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.state.MoveSelection(-1) {
				return m, emitPreviewSelection(m.state.SelectedItem().Name)
			}
			return m, nil
		case tea.KeyDown:
			if m.state.MoveSelection(1) {
				return m, emitPreviewSelection(m.state.SelectedItem().Name)
			}
			return m, nil
		case tea.KeyEnter:
			return m, emitConfirmSelection(m.state.SelectedItem().Name)
		case tea.KeyEsc:
			if m.input.Value() != "" {
				m.input.SetValue("")
				m.state.SetQuery("")
				return m, nil
			}
			return m, emitCancelSelection()
		case tea.KeyTab:
			m.state.ToggleViewMode()
			return m, nil
		case tea.KeyBackspace:
			before := m.state.SelectedItem().Name
			m.input, _ = m.input.Update(msg)
			m.state.SetQuery(m.input.Value())
			if selected := m.state.SelectedItem().Name; selected != "" && selected != before {
				return m, emitPreviewSelection(selected)
			}
			return m, nil
		case tea.KeyRunes:
			if len(msg.Runes) == 1 && msg.Runes[0] == 'f' {
				m.state.ToggleFavorite()
				if m.effects.onSaveFavorites != nil {
					if err := m.effects.onSaveFavorites(m.state.FavoriteNames()); err != nil {
						m.status = err.Error()
					}
				}
				return m, nil
			}
			before := m.state.SelectedItem().Name
			m.input, _ = m.input.Update(msg)
			m.state.SetQuery(m.input.Value())
			if selected := m.state.SelectedItem().Name; selected != "" && selected != before {
				return m, emitPreviewSelection(selected)
			}
			return m, nil
		}
	}

	m.input, _ = m.input.Update(msg)
	m.state.SetQuery(m.input.Value())
	return m, nil
}

func (m pickerModel) View() string {
	return renderPickerView(m)
}

func emitPreviewSelection(name string) tea.Cmd {
	return func() tea.Msg {
		if name == "" {
			return nil
		}
		return previewSelectionMsg{schemeName: name}
	}
}

func emitConfirmSelection(name string) tea.Cmd {
	return func() tea.Msg {
		return confirmSelectionMsg{schemeName: name}
	}
}

func emitCancelSelection() tea.Cmd {
	return func() tea.Msg {
		return cancelSelectionMsg{}
	}
}

func (m pickerModel) selectedSchemeTitle() string {
	item := m.state.SelectedItem()
	if item.Name == "" {
		if m.previewName != "" {
			return fmt.Sprintf("previewing %s", m.previewName)
		}
		return "no theme selected"
	}
	return fmt.Sprintf("previewing %s", item.Name)
}
