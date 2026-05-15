package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dotpablos/vibel/internal/models"
	spotifyapi "github.com/zmb3/spotify/v2"
)

type screen int

const (
	screenLogin screen = iota
	screenSearch
	screenPlayer
	screenQueue
)

type appModel struct {
	client *spotifyapi.Client

	currentScreen screen

	login  *models.LoginModel
	search *screenModel
	player *screenModel
	queue  *screenModel
}

func NewAppModel() *appModel {
	return &appModel{
		client:        &spotifyapi.Client{},
		currentScreen: screenLogin,
		login:         &models.LoginModel{},
		search:        newScreenModel("Search"),
		player:        newScreenModel("Player"),
		queue:         newScreenModel("Queue"),
	}
}

func (m *appModel) Init() tea.Cmd {
	return nil
}

func (m *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "1":
			m.currentScreen = screenLogin
		case "2":
			m.currentScreen = screenSearch
		case "3":
			m.currentScreen = screenPlayer
		case "4":
			m.currentScreen = screenQueue
		}
	}

	switch m.currentScreen {
	case screenLogin:
		return m, m.login.Update(msg)
	case screenSearch:
		return m, m.search.Update(msg)
	case screenPlayer:
		return m, m.player.Update(msg)
	case screenQueue:
		return m, m.queue.Update(msg)
	default:
		return m, nil
	}
}

func (m *appModel) View() string {
	switch m.currentScreen {
	case screenLogin:
		return m.login.View()
	case screenSearch:
		return m.search.View()
	case screenPlayer:
		return m.player.View()
	case screenQueue:
		return m.queue.View()
	default:
		return "Unknown screen"
	}
}
