package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dotpablos/vibel/internal/cli"
	"github.com/dotpablos/vibel/internal/tui"
)

func main() {
	userArgs := os.Args[1:]

	// TUI Entry
	if len(userArgs) == 0 {
		p := tea.NewProgram(tui.NewAppModel())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Something went wrong when starting the TUI: %v\n", err)
			os.Exit(1)
		}
	}

	// CLI entry
	if err := cli.RunCLI(userArgs); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
