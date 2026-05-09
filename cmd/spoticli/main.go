package main

import (
	"fmt"
	"os"

	"github.com/dotpablos/spoticli/internal/cli"
)

func main() {
	userArgs := os.Args[1:]

	// TUI Entry
	if len(userArgs) == 0 {
		os.Exit(0)
	}

	// CLI entry
	if err := cli.RunCLI(userArgs); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
