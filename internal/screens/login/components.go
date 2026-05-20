package login

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dotpablos/vibel/internal/assets"
	"github.com/dotpablos/vibel/internal/ui"
	"github.com/dotpablos/vibel/internal/version"
)

func (m *Model) window(width int) string {
	m.ensureCommands()

	displayURL := m.authURL
	if displayURL == "" {
		displayURL = "press r to generate a login link"
	}

	bodyWidth := width - 4
	dividerWidth := 1
	leftWidth := (bodyWidth - dividerWidth) / 2
	rightWidth := bodyWidth - leftWidth - dividerWidth
	rightPadding := 4
	panelWidth := rightWidth - rightPadding - 2
	if panelWidth < 1 {
		panelWidth = rightWidth
	}

	notice := lipgloss.NewStyle().
		Width(panelWidth).
		Padding(1, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(ui.SoftGold).
		Render(
			lipgloss.NewStyle().Foreground(ui.Gold).Bold(true).Render("<!> you're not authenticated") + "\n" +
				lipgloss.NewStyle().Foreground(ui.Dim).Render(
					"run the command below to open spotify's oauth page in your browser, then return here once\n"+
						"authorised. or paste the url manually into any browser.",
				),
		)

	label := lipgloss.NewStyle().Foreground(ui.Dim)

	command := lipgloss.NewStyle().
		Width(panelWidth).
		Padding(0, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(ui.Line).
		Render(
			lipgloss.NewStyle().Foreground(ui.Gold).Render("$") +
				" " +
				lipgloss.NewStyle().Foreground(ui.BrightText).Bold(true).Render("vibel login"),
		)

	url := lipgloss.NewStyle().
		Width(panelWidth).
		Padding(1, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(ui.Line).
		Render(lipgloss.NewStyle().Foreground(ui.BrightText).Italic(true).Render(displayURL))

	sections := []string{
		notice,
		"",
		label.Render("COMMAND"),
		command,
		"",
		label.Render("MANUAL URL"),
		url,
	}

	if status := renderStatusPanel(m.status, m.statusOK, panelWidth); status != "" {
		sections = append(sections, "", label.Render("STATUS"), status)
	}

	right := lipgloss.JoinVertical(lipgloss.Left, sections...)

	left := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().Render(strings.TrimSuffix(assets.GetLogoString(), "\n")),
		"",
		lipgloss.NewStyle().Foreground(ui.Dim).Render("VIBEL"),
		lipgloss.NewStyle().Foreground(ui.Dim).Render("v"+version.Release),
	)

	contentHeight := max(lipgloss.Height(left), lipgloss.Height(right))
	divider := lipgloss.NewStyle().
		Foreground(ui.Dim).
		Render(strings.TrimSuffix(strings.Repeat("│\n", contentHeight), "\n"))

	bodyContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.Place(leftWidth, contentHeight, lipgloss.Center, lipgloss.Center, left),
		divider,
		lipgloss.NewStyle().PaddingLeft(rightPadding).Render(
			lipgloss.Place(rightWidth-rightPadding, contentHeight, lipgloss.Left, lipgloss.Top, right),
		),
	)

	body := lipgloss.NewStyle().
		Padding(0, 2, 1, 2).
		Render(lipgloss.PlaceHorizontal(bodyWidth, lipgloss.Left, bodyContent))

	return lipgloss.NewStyle().
		MarginTop(1).
		MarginLeft(2).
		Render(lipgloss.JoinVertical(lipgloss.Left, body))
}

func renderStatusPanel(status string, ok bool, width int) string {
	if status == "" {
		return ""
	}

	color := ui.Green
	if !ok {
		color = ui.Red
	}

	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(color).
		Render(lipgloss.NewStyle().Foreground(color).Render(status))
}
