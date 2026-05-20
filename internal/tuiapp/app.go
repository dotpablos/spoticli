package tuiapp

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dotpablos/vibel/internal/app"
	"github.com/dotpablos/vibel/internal/screens/login"
	"github.com/dotpablos/vibel/internal/screens/player"
	"github.com/dotpablos/vibel/internal/screens/queue"
	"github.com/dotpablos/vibel/internal/screens/search"
	"github.com/dotpablos/vibel/internal/spotify"
)

type screenModel interface {
	Update(tea.Msg) tea.Cmd
	View() string
}

type Model struct {
	currentScreen app.Screen
	screens       map[app.Screen]screenModel
	services      spotify.Services
}

func New() *Model {
	services := spotify.NewServices()

	return &Model{
		currentScreen: app.ScreenLogin,
		screens: map[app.Screen]screenModel{
			app.ScreenLogin:  login.New(services.Auth),
			app.ScreenSearch: search.New(),
			app.ScreenPlayer: player.New(services.Playback),
			app.ScreenQueue:  queue.New(),
		},
		services: services,
	}
}

func (m *Model) Init() tea.Cmd {
	if m.services.Auth.HasSavedSession() {
		return app.Navigate(app.ScreenPlayer)
	}

	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case app.NavigateMsg:
		m.setScreen(msg.Screen)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		default:
			if screen, ok := app.ScreenForKey(msg.String()); ok {
				m.setScreen(screen)
			}
		}
	}

	active := m.activeScreen()
	if active == nil {
		return m, nil
	}

	return m, active.Update(msg)
}

func (m *Model) View() string {
	active := m.activeScreen()
	if active == nil {
		return "Unknown screen"
	}

	return active.View()
}

func (m *Model) activeScreen() screenModel {
	return m.screens[m.currentScreen]
}

func (m *Model) setScreen(screen app.Screen) {
	if _, ok := m.screens[screen]; ok {
		m.currentScreen = screen
	}
}
