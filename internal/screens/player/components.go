package player

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dotpablos/vibel/internal/spotify"
	"github.com/dotpablos/vibel/internal/ui"
)

const queuePanelBackground = lipgloss.Color("#1B1B1B")

func (m *Model) renderNowPlayingPanel(width int, height int) string {
	commandLog := renderCommandLogMeta(m.status, m.statusOK)

	return renderPlayerPanel(
		"now playing",
		commandLog,
		width,
		height,
		"",
		m.renderNowPlayingBody(max(width-2, 1), max(height-4, 1)),
		lipgloss.Center,
	)
}

func (m *Model) renderQueuePanel(width int, height int) string {
	queue := m.queueEntries()
	if len(queue) == 0 {
		queue = demoQueueEntries()
	}

	return renderPlayerPanel(
		"queue",
		strconv.Itoa(len(queue))+" tracks",
		width,
		height,
		queuePanelBackground,
		renderQueueBody(queue, max(width-2, 1), max(height-4, 1)),
		lipgloss.Top,
	)
}

func renderPlayerPanel(title string, meta string, width int, height int, background lipgloss.Color, body string, bodyVAlign lipgloss.Position) string {
	innerWidth := max(width-2, 1)
	innerHeight := max(height-2, 1)

	header := renderPlayerPanelHeader(title, meta, innerWidth)
	separator := lipgloss.NewStyle().
		Foreground(ui.Dim).
		Render(strings.Repeat("─", innerWidth))

	bodyHeight := innerHeight - lipgloss.Height(header) - lipgloss.Height(separator)
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	bodyStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Height(bodyHeight)
	if background != "" {
		bodyStyle = bodyStyle.Background(background)
	}

	bodyBlock := bodyStyle.Render(lipgloss.Place(innerWidth, bodyHeight, lipgloss.Left, bodyVAlign, body))

	panelStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Height(innerHeight).
		Border(lipgloss.NormalBorder()).
		BorderForeground(ui.Dim)
	if background != "" {
		panelStyle = panelStyle.Background(background)
	}

	return lipgloss.NewStyle().
		Render(panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, separator, bodyBlock)))
}

func renderPlayerPanelHeader(title string, meta string, width int) string {
	left := lipgloss.NewStyle().
		Foreground(ui.Gold).
		Bold(true).
		Render("• " + strings.ToUpper(title))

	right := lipgloss.NewStyle().
		Foreground(ui.Dim).
		Render(meta)

	return lipgloss.NewStyle().
		Width(width).
		Render(ui.SpaceBetween(left, right, width))
}

func (m *Model) renderNowPlayingBody(width int, height int) string {
	artHeight := min(9, max(7, height/2))
	if height < 16 {
		artHeight = min(7, max(5, height/3))
	}

	current := m.playback
	if !current.HasCurrentTrack() {
		current = demoPlayback
	}

	sections := []string{
		renderAlbumArt(width, artHeight),
		lipgloss.NewStyle().Foreground(ui.BrightText).Bold(true).Render(current.Current.Title),
		lipgloss.NewStyle().Foreground(ui.Gold).Bold(true).Render(current.Current.Artist),
		lipgloss.NewStyle().Foreground(ui.Dim).Render(current.Current.Album),
		renderPlayerProgressTimes(width, current.ProgressMs, current.Current.DurationMs),
		renderMeter(width, progressRatio(current.ProgressMs, current.Current.DurationMs), ui.Gold, ui.Line),
		m.renderControlRows(width),
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func renderAlbumArt(width int, height int) string {
	label := lipgloss.NewStyle().
		Foreground(ui.Dim).
		Render("FLAC · 44.1kHz")

	graphicHeight := max(height-1, 1)
	graphic := lipgloss.NewStyle().
		Foreground(ui.Dim).
		Render(renderRecordGraphic())

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.Place(width, graphicHeight, lipgloss.Center, lipgloss.Center, graphic),
		lipgloss.PlaceHorizontal(width, lipgloss.Right, label),
	)

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.NormalBorder()).
		BorderForeground(ui.Dim).
		Render(content)
}

func renderRecordGraphic() string {
	lines := []string{
		"    .--------.    ",
		"  .'  .----.  '.  ",
		" /  .'      '.  \\ ",
		"|  |   .--.   |  |",
		"|  |  | " + lipgloss.NewStyle().Foreground(ui.Gold).Bold(true).Render("o") + "  |  |  |",
		"|  |   '--'   |  |",
		" \\  '.      .'  / ",
		"  '.  '----'  .'  ",
		"    '--------'    ",
	}

	return strings.Join(lines, "\n")
}

func renderPlayerProgressTimes(width int, progressMs int, durationMs int) string {
	left := lipgloss.NewStyle().Foreground(ui.Dim).Render(spotify.FormatDuration(progressMs))
	right := lipgloss.NewStyle().Foreground(ui.Dim).Render(spotify.FormatDuration(durationMs))

	return ui.SpaceBetween(left, right, width)
}

func (m *Model) renderControlRows(width int) string {
	buttons := []controlButton{
		m.toggleButton(),
		{icon: ">>|", label: "skip"},
	}

	return renderControlButtonRow(width, buttons)
}

func renderCommandLogMeta(status string, ok bool) string {
	if status == "" {
		return ""
	}

	prefixColor := ui.Green
	prefix := "log"
	if !ok {
		prefixColor = ui.Red
		prefix = "err"
	}

	return lipgloss.NewStyle().
		Foreground(prefixColor).
		Render(prefix) + lipgloss.NewStyle().Foreground(ui.Dim).Render(": "+status)
}

func renderControlButtonRow(width int, buttons []controlButton) string {
	gap := 1
	totalGap := gap * (len(buttons) - 1)
	buttonWidth := max((width-totalGap)/len(buttons), 8)

	parts := make([]string, 0, len(buttons))
	for index, button := range buttons {
		if index > 0 {
			parts = append(parts, " ")
		}
		parts = append(parts, renderControlButton(button, buttonWidth))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func renderControlButton(button controlButton, width int) string {
	innerWidth := max(width-2, 4)
	accent := ui.BrightText
	border := ui.Dim
	if button.accent != "" {
		accent = button.accent
		border = button.accent
	}

	icon := lipgloss.NewStyle().
		Width(innerWidth).
		Align(lipgloss.Center).
		Foreground(accent).
		Bold(true).
		Render(button.icon)

	label := lipgloss.NewStyle().
		Width(innerWidth).
		Align(lipgloss.Center).
		Foreground(accent).
		Bold(true).
		Render(ui.Truncate(button.label, innerWidth))

	return lipgloss.NewStyle().
		Width(innerWidth).
		Border(lipgloss.NormalBorder()).
		BorderForeground(border).
		Render(icon + "\n" + label)
}

func renderMeter(width int, ratio float64, fill lipgloss.Color, empty lipgloss.Color) string {
	if width < 1 {
		width = 1
	}

	filled := min(width, max(int(float64(width)*ratio), 0))

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.NewStyle().Foreground(fill).Render(strings.Repeat("━", filled)),
		lipgloss.NewStyle().Foreground(empty).Render(strings.Repeat("━", width-filled)),
	)
}

func renderQueueBody(queue []queueEntry, width int, height int) string {
	maxTracks := max(height/2, 1)
	if maxTracks > len(queue) {
		maxTracks = len(queue)
	}

	rows := make([]string, 0, maxTracks)
	for _, track := range queue[:maxTracks] {
		rows = append(rows, renderQueueTrack(track, width))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderQueueTrack(track queueEntry, width int) string {
	prefix := lipgloss.NewStyle().Foreground(ui.Dim).Render(" ")
	titleStyle := lipgloss.NewStyle().Foreground(ui.Text)
	if track.IsCurrent {
		prefix = lipgloss.NewStyle().Foreground(ui.Gold).Bold(true).Render("▶")
		titleStyle = lipgloss.NewStyle().Foreground(ui.BrightText).Bold(true)
	} else if track.Position > 0 {
		prefix = lipgloss.NewStyle().
			Foreground(ui.Line).
			Width(2).
			Align(lipgloss.Right).
			Render(strconv.Itoa(track.Position))
	}

	durationText := spotify.FormatDuration(track.Track.DurationMs)
	leftWidth := max(width-lipgloss.Width(durationText)-3, 8)
	title := titleStyle.Render(ui.Truncate(track.Track.Title, leftWidth))
	duration := lipgloss.NewStyle().
		Foreground(ui.Dim).
		Render(durationText)

	lineOne := ui.SpaceBetween(
		lipgloss.JoinHorizontal(lipgloss.Top, prefix, " ", title),
		duration,
		width,
	)

	lineTwo := lipgloss.NewStyle().
		Foreground(ui.Dim).
		Render(strings.Repeat(" ", 2) + ui.Truncate(track.Track.Artist, width-2))

	return lineOne + "\n" + lineTwo
}
