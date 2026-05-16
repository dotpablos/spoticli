package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dotpablos/vibel/internal/assets"
)

const (
	defaultLoginWidth  = 96
	minLoginWidth      = 62
	defaultLoginHeight = 24
)

type LoginModel struct {
	width  int
	height int
}

func (m *LoginModel) Update(msg tea.Msg) tea.Cmd {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = size.Width
		m.height = size.Height
	}

	return nil
}

func (m *LoginModel) View() string {
	width := m.contentWidth()
	innerWidth := width - 4

	window := m.window(innerWidth)
	commandbar := m.commandbar(innerWidth)
	content := lipgloss.JoinVertical(lipgloss.Left, window, commandbar)

	return lipgloss.Place(
		width,
		m.contentHeight(),
		lipgloss.Left,
		lipgloss.Center,
		lipgloss.NewStyle().
			Foreground(Text).
			Padding(0, 1).
			Render(content),
	)
}

func (m *LoginModel) contentWidth() int {
	if m.width <= 0 {
		return defaultLoginWidth
	}

	// The app shell already spends a little room on its own border.
	width := m.width - 6
	if width < minLoginWidth {
		return minLoginWidth
	}

	return width
}

func (m *LoginModel) contentHeight() int {
	if m.height <= 0 {
		return defaultLoginHeight
	}

	// Match the border space reserved by the app shell.
	height := m.height - 2
	if height < 1 {
		return 1
	}

	return height
}

func (m *LoginModel) window(width int) string {
	windowWidth := width - 4
	bodyWidth := windowWidth - 4
	logoContent := strings.TrimSuffix(assets.GetLogoString(), "\n")
	logoWidth := lipgloss.Width(logoContent)
	contentWidth := bodyWidth - logoWidth - 4
	sideBySide := contentWidth >= 48
	if !sideBySide {
		contentWidth = bodyWidth - 4
	}

	logo := lipgloss.NewStyle().Render(logoContent)

	notice := lipgloss.NewStyle().
		Width(contentWidth).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SoftGold).
		Render(
			lipgloss.NewStyle().Foreground(Gold).Bold(true).Render("not authenticated") + "\n\n" +
				lipgloss.NewStyle().Foreground(Dim).Render(
					"run the command below to open spotify's oauth page in your browser, then return here once\n"+
						"authorised. or paste the url manually into any browser.",
				),
		)

	command := lipgloss.NewStyle().
		Width(contentWidth).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		Render(
			lipgloss.NewStyle().Foreground(Gold).Render("$") +
				" " +
				lipgloss.NewStyle().Foreground(BrightText).Bold(true).Render("spotifytui auth --open-browser"),
		)

	url := lipgloss.NewStyle().
		Width(contentWidth).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SoftGold).
		Render(
			lipgloss.NewStyle().Foreground(Dim).Render("MANUAL URL") + "\n\n" +
				lipgloss.NewStyle().Foreground(Gold).Bold(true).Render(
					"https://accounts.spotify.com/authorize?client_id=f3b8c2a1d04e&response_type=code&redirect_uri\n"+
						"=http%3A%2F%2Flocalhost%3A8888%2Fcallback&scope=user-read-playback-state+user-modify-playback\n"+
						"-state",
				),
		)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		notice,
		"",
		command,
		"",
		url,
	)

	bodyContent := lipgloss.JoinHorizontal(lipgloss.Center, logo, lipgloss.NewStyle().Width(4).Render(""), content)
	if !sideBySide {
		bodyContent = lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.PlaceHorizontal(bodyWidth, lipgloss.Center, logo),
			"",
			content,
		)
	}

	body := lipgloss.NewStyle().
		Padding(0, 2, 1, 2).
		Render(lipgloss.PlaceHorizontal(
			bodyWidth,
			lipgloss.Left,
			bodyContent,
		))

	return lipgloss.NewStyle().
		MarginTop(1).
		MarginLeft(2).
		Render(lipgloss.JoinVertical(lipgloss.Left, body))
}

func (m *LoginModel) commandbar(width int) string {
	controls := lipgloss.NewStyle().
		Foreground(Dim).
		Render(
			lipgloss.NewStyle().Foreground(Gold).Bold(true).Render("enter") + " open browser   ·   " +
				lipgloss.NewStyle().Foreground(Gold).Bold(true).Render("y") + " copy url   ·   " +
				lipgloss.NewStyle().Foreground(Gold).Bold(true).Render("r") + " regenerate token   ·   " +
				lipgloss.NewStyle().Foreground(Gold).Bold(true).Render(":q") + " quit",
		)

	return lipgloss.NewStyle().
		MarginLeft(2).
		Render(inlineRule(width-4, "╰", "╯", "─", " "+controls+" "))
}

func inlineRule(width int, left, right, fill, content string) string {
	innerWidth := max(width-lipgloss.Width(left)-lipgloss.Width(right), 0)
	contentWidth := lipgloss.Width(content)
	if contentWidth >= innerWidth {
		return left + content + right
	}

	fillWidth := innerWidth - contentWidth
	leftFill := fillWidth / 2
	rightFill := fillWidth - leftFill

	leftBorder := lipgloss.NewStyle().Foreground(Dim).Render(left + strings.Repeat(fill, leftFill))
	rightBorder := lipgloss.NewStyle().Foreground(Dim).Render(strings.Repeat(fill, rightFill) + right)

	return leftBorder +
		content + rightBorder
}
