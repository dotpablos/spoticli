package models

import tea "github.com/charmbracelet/bubbletea"

type LoginModel struct{}

func (m *LoginModel) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (m *LoginModel) View() string {
	return ""
}

