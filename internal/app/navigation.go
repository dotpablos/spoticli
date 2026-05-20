package app

import tea "github.com/charmbracelet/bubbletea"

type Screen string

const (
	ScreenLogin  Screen = "login"
	ScreenSearch Screen = "search"
	ScreenPlayer Screen = "player"
	ScreenQueue  Screen = "queue"
)

type NavigateMsg struct {
	Screen Screen
}

func Navigate(screen Screen) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{Screen: screen}
	}
}

func ScreenForKey(key string) (Screen, bool) {
	switch key {
	case "1":
		return ScreenLogin, true
	case "2":
		return ScreenSearch, true
	case "3":
		return ScreenPlayer, true
	case "4":
		return ScreenQueue, true
	default:
		return "", false
	}
}
