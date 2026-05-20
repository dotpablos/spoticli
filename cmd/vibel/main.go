package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dotpablos/vibel/internal/cli"
	"github.com/dotpablos/vibel/internal/tuiapp"
)

func main() {
	userArgs := os.Args[1:]

	if len(userArgs) == 0 {
		program := tea.NewProgram(tuiapp.New(), tea.WithAltScreen())
		if _, err := program.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Something went wrong when starting the TUI: %v\n", err)
			os.Exit(1)
		}
	}

	if err := cli.RunCLI(userArgs); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
