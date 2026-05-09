package cli

import (
	"fmt"
	"strconv"
)

type command struct {
	name               string
	excuseFromHelpMenu bool
	description        string
	run                func(args []string) error
}

var commands = map[string]command{
	"pause": {
		name:               "pause",
		description:        "Pause the current Spotify track",
		excuseFromHelpMenu: false,
		run: func(args []string) error {
			return requireNoArgs("pause", args)
		},
	},
	"play": {
		name:               "play",
		description:        "Resume the current Spotify track",
		excuseFromHelpMenu: false,
		run: func(args []string) error {
			return requireNoArgs("play", args)
		},
	},
	"forward": {
		name:               "forward",
		description:        "Skip to the next track in the queue",
		excuseFromHelpMenu: false,
		run: func(args []string) error {
			return requireNoArgs("forward", args)
		},
	},
	"back": {
		name:               "back",
		description:        "Return to the previous track",
		excuseFromHelpMenu: false,
		run: func(args []string) error {
			return requireNoArgs("back", args)
		},
	},
	"login": {
		name:               "login",
		description:        "Authenticate with Spotify",
		excuseFromHelpMenu: false,
		run: func(args []string) error {
			return requireNoArgs("login", args)
		},
	},
	"logout": {
		name:               "logout",
		description:        "Clear the current Spotify session",
		excuseFromHelpMenu: false,
		run: func(args []string) error {
			return requireNoArgs("logout", args)
		},
	},
	"song": {
		name:               "song",
		description:        "Show the current song name and playback time",
		excuseFromHelpMenu: false,
		run: func(args []string) error {
			return requireNoArgs("song", args)
		},
	},
	"list": {
		name:               "list",
		description:        "List the next <count> songs in the queue",
		excuseFromHelpMenu: false,
		run: func(args []string) error {
			return requirePositiveCountArg(args)
		},
	},
}
var helpCmd = command{
	name:               "help",
	description:        "",
	excuseFromHelpMenu: true,
	run: func(args []string) error {
		fmt.Println(`
SPOTICLI
In order to launch the TUI, just enter "spoticli" with no additional flags.

USAGE: spoticli [COMMAND] [AMOUNT (if needed)]

Note: The CLI is meant to be used to interface with Spotify for simple actions.
Commands cannot be chained.
For a more in-depth experience, use the TUI.

COMMANDS:`)

		for _, cmd := range commands {
			if !cmd.excuseFromHelpMenu {
				fmt.Println("\t\t" + cmd.name + ": \n\t\t\t" + cmd.description)
			}
		}
		return nil
	},
}

func RunCLI(args []string) error {
	if len(args) == 0 {
		return nil
	}
	// Build help commands at RT
	commands["--help"] = helpCmd
	commands["-h"] = helpCmd

	cmd, ok := commands[args[0]]
	if !ok {
		return fmt.Errorf("SPOTICLI unknown command %q, try using -h or --help to see a list of commands", args[0])
	}

	return cmd.run(args[1:])
}

func requireNoArgs(name string, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: spoticli %s", name)
	}

	return nil
}

func requirePositiveCountArg(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: spoticli list <count>")
	}

	count, err := strconv.Atoi(args[0])
	if err != nil || count < 1 {
		return fmt.Errorf("list count must be a positive integer")
	}

	_ = count
	return nil
}
