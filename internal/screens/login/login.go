package login

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dotpablos/vibel/internal/app"
	"github.com/dotpablos/vibel/internal/spotify"
	"github.com/dotpablos/vibel/internal/ui"
)

type Model struct {
	app.ScreenLayout
	auth     *spotify.AuthService
	authURL  string
	login    *spotify.LoginSession
	commands []app.Command
	status   string
	statusOK bool
}

type loginCommandResultMsg struct {
	status string
	err    error
}

type loginCompletedMsg struct {
	session *spotify.LoginSession
	err     error
}

func New(auth *spotify.AuthService) *Model {
	if auth == nil {
		auth = spotify.NewAuthService()
	}

	return &Model{auth: auth}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.ensureCommands()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Resize(msg)
	case loginCompletedMsg:
		if msg.session != m.login {
			return nil
		}
		if msg.err != nil {
			m.setStatus(msg.err.Error(), false)
			return nil
		}

		return app.Navigate(app.ScreenPlayer)
	case loginCommandResultMsg:
		if msg.err != nil {
			m.setStatus(msg.err.Error(), false)
			return nil
		}

		if msg.status != "" {
			m.setStatus(msg.status, true)
		}
	case tea.KeyMsg:
		return app.CommandForKey(msg.String(), m.commands)
	}

	return nil
}

func (m *Model) View() string {
	m.ensureCommands()

	width := m.ContentWidth()
	innerWidth := width - 4

	window := m.window(innerWidth)
	commandBar := app.RenderCommandBar(m.commands, innerWidth, "╰", "╯", "─")
	content := lipgloss.JoinVertical(lipgloss.Left, window, commandBar)

	return lipgloss.Place(
		width,
		m.ContentHeight(),
		lipgloss.Left,
		lipgloss.Center,
		lipgloss.NewStyle().
			Foreground(ui.Text).
			Padding(0, 1).
			Render(content),
	)
}
