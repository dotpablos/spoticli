package search

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) handleWindowSize(msg tea.Msg) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.Resize(size)
	}
}
