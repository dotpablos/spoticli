package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dotpablos/vibel/internal/ui"
)

type Command struct {
	Name string
	Key  string
	Exec func() tea.Cmd
}

func GlobalCommands() []Command {
	return []Command{
		{
			Name: "Exit",
			Key:  "ctrl+c",
			Exec: func() tea.Cmd {
				return tea.Quit
			},
		},
	}
}

func CommandForKey(key string, commands []Command) tea.Cmd {
	for _, command := range commands {
		if key == command.Key {
			return command.Exec()
		}
	}

	return nil
}

func RenderCommandBar(commands []Command, width int, leftEdge string, rightEdge string, fill string) string {
	parts := make([]string, 0, len(commands))

	keyStyle := lipgloss.NewStyle().
		Foreground(ui.Gold).
		Bold(true)

	textStyle := lipgloss.NewStyle().
		Foreground(ui.Dim)

	for _, cmd := range commands {
		part := keyStyle.Render(cmd.Key) + " " + textStyle.Render(cmd.Name)
		parts = append(parts, part)
	}

	controls := strings.Join(parts, "  ·  ")

	innerWidth := max(width-lipgloss.Width(leftEdge)-lipgloss.Width(rightEdge), 0)
	contentWidth := lipgloss.Width(controls)
	if contentWidth >= innerWidth {
		return leftEdge + controls + rightEdge
	}

	fillWidth := innerWidth - contentWidth
	leftFill := fillWidth / 2
	rightFill := fillWidth - leftFill

	leftBorder := lipgloss.NewStyle().Foreground(ui.Dim).Render(leftEdge + strings.Repeat(fill, leftFill))
	rightBorder := lipgloss.NewStyle().Foreground(ui.Dim).Render(strings.Repeat(fill, rightFill) + rightEdge)

	return lipgloss.NewStyle().
		MarginLeft(2).
		Render(leftBorder + controls + rightBorder)
}
