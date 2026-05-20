package app

import tea "github.com/charmbracelet/bubbletea"

type ScreenLayout struct {
	width  int
	height int
}

const (
	defaultScreenWidth  = 96
	minScreenWidth      = 62
	defaultScreenHeight = 24
)

func (m *ScreenLayout) Resize(size tea.WindowSizeMsg) {
	m.width = size.Width
	m.height = size.Height
}

func (m *ScreenLayout) ContentWidth() int {
	if m.width <= 0 {
		return defaultScreenWidth
	}

	width := m.width - 6
	if width < minScreenWidth {
		return minScreenWidth
	}

	return width
}

func (m *ScreenLayout) ContentHeight() int {
	if m.height <= 0 {
		return defaultScreenHeight
	}

	height := m.height - 2
	if height < 1 {
		return 1
	}

	return height
}
