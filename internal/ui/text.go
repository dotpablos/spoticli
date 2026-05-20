package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func SpaceBetween(left string, right string, width int) string {
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)

	if leftWidth+rightWidth >= width {
		if right == "" {
			return Truncate(left, width)
		}

		available := max(width-rightWidth-1, 1)
		return Truncate(left, available) + " " + right
	}

	return left + strings.Repeat(" ", width-leftWidth-rightWidth) + right
}

func Truncate(value string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= width {
		return value
	}

	if width <= 3 {
		return string(runes[:width])
	}

	return string(runes[:width-3]) + "..."
}
