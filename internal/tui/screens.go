package tui

import tea "github.com/charmbracelet/bubbletea"

// screenModel is a temporary stand-in until each screen gets its own model.
type screenModel struct {
	title string
}

func newScreenModel(title string) *screenModel {
	return &screenModel{title: title}
}

func (m *screenModel) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (m *screenModel) View() string {
	return m.title
}
