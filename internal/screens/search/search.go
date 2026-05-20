package search

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dotpablos/vibel/internal/app"
)

type Model struct {
	app.ScreenLayout
}

func New() *Model {
	return &Model{}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.handleWindowSize(msg)
	return nil
}
