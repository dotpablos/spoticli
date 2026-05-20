package login

import (
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dotpablos/vibel/internal/app"
	"github.com/dotpablos/vibel/internal/spotify"
)

func (m *Model) ensureCommands() {
	if len(m.commands) == 0 {
		m.commands = m.updateCommands()
	}
}

func (m *Model) setStatus(status string, ok bool) {
	m.status = status
	m.statusOK = ok
}

func (m *Model) emitResult(status string, err error) tea.Cmd {
	return func() tea.Msg {
		return loginCommandResultMsg{status: status, err: err}
	}
}

func (m *Model) waitForLoginCmd(session *spotify.LoginSession) tea.Cmd {
	return func() tea.Msg {
		return loginCompletedMsg{session: session, err: session.Wait()}
	}
}

func (m *Model) beginLogin(openBrowser bool) tea.Cmd {
	login, err := m.auth.BeginLogin()
	if err != nil {
		return m.emitResult("", err)
	}

	if m.login != nil {
		m.login.Close()
	}

	m.login = login
	m.authURL = login.URL
	m.commands = m.updateCommands()
	m.setStatus("", true)

	cmds := []tea.Cmd{m.waitForLoginCmd(login)}
	if openBrowser {
		if err := login.Open(); err != nil {
			cmds = append(cmds, m.emitResult("", err))
		} else {
			cmds = append(cmds, m.emitResult("opened browser for Spotify login", nil))
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) updateCommands() []app.Command {
	globalCommands := app.GlobalCommands()
	defaults := []app.Command{
		{
			Name: "re-generate login link",
			Key:  "r",
			Exec: func() tea.Cmd {
				return m.beginLogin(false)
			},
		},
		{
			Name: "open browser at url",
			Key:  "o",
			Exec: func() tea.Cmd {
				if m.login != nil {
					return m.emitResult("opened browser for Spotify login", m.login.Open())
				}

				return m.beginLogin(true)
			},
		},
	}

	if m.authURL != "" {
		defaults = append(defaults, app.Command{
			Name: "copy login link",
			Key:  "c",
			Exec: func() tea.Cmd {
				return m.emitResult("copied login link", clipboard.WriteAll(m.authURL))
			},
		})
	}

	commands := make([]app.Command, 0, len(globalCommands)+len(defaults))
	commands = append(commands, globalCommands...)
	commands = append(commands, defaults...)
	return commands
}
