package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dotpablos/vibel/internal/spotify"
	"github.com/pkg/browser"
)

type command struct {
	name         string
	aliases      []string
	description  string
	hideFromHelp bool
	run          func(args []string, services spotify.Services) error
}

func (cmd command) matches(name string) bool {
	if cmd.name == name {
		return true
	}

	for _, alias := range cmd.aliases {
		if alias == name {
			return true
		}
	}

	return false
}

func (cmd command) execute(args []string) error {
	services := spotify.NewServices()

	if err := cmd.run(args, services); err != nil {
		return fmt.Errorf("%s: %w", cmd.name, err)
	}

	return nil
}

func commandList() []command {
	return []command{
		{
			name:        "help",
			aliases:     []string{"-h", "--help"},
			description: "Show available commands",
			run:         runHelp,
		},
		{
			name:        "login",
			description: "Authenticate with Spotify",
			run: func(args []string, services spotify.Services) error {
				if err := requireNoArgs("login", args); err != nil {
					return err
				}

				return services.Auth.Login(true)
			},
		},
		{
			name:        "logout",
			description: "Clear the current Spotify session",
			run: func(args []string, services spotify.Services) error {
				if err := requireNoArgs("logout", args); err != nil {
					return err
				}

				return services.Auth.Logout()
			},
		},
		{
			name:        "pause",
			description: "Pause the current Spotify track",
			run: func(args []string, services spotify.Services) error {
				if err := requireNoArgs("pause", args); err != nil {
					return err
				}

				return services.Playback.Pause(context.Background())
			},
		},
		{
			name:        "play",
			description: "Resume the current Spotify track",
			run: func(args []string, services spotify.Services) error {
				if err := requireNoArgs("play", args); err != nil {
					return err
				}

				return services.Playback.Play(context.Background())
			},
		},
		{
			name:        "forward",
			description: "Skip to the next track in the queue",
			run: func(args []string, services spotify.Services) error {
				if err := requireNoArgs("forward", args); err != nil {
					return err
				}

				return services.Playback.Next(context.Background())
			},
		},
		{
			name:        "back",
			description: "Return to the previous track",
			run: func(args []string, services spotify.Services) error {
				if err := requireNoArgs("back", args); err != nil {
					return err
				}

				return services.Playback.Previous(context.Background())
			},
		},
		{
			name:        "song",
			description: "Show the current song name and playback time",
			run: func(args []string, services spotify.Services) error {
				if err := requireNoArgs("song", args); err != nil {
					return err
				}

				playback, err := services.Playback.Snapshot(context.Background())
				if err != nil {
					return err
				}

				if !playback.HasCurrentTrack() {
					fmt.Println("Nothing is currently playing.")
					return nil
				}

				fmt.Printf(
					"%s [%s / %s]\n",
					spotify.FormatTrack(playback.Current),
					spotify.FormatDuration(playback.ProgressMs),
					spotify.FormatDuration(playback.Current.DurationMs),
				)
				return nil
			},
		},
		{
			name:        "list",
			description: "List the next <count> songs in the queue",
			run: func(args []string, services spotify.Services) error {
				count, err := requirePositiveCountArg(args)
				if err != nil {
					return err
				}

				playback, err := services.Playback.Snapshot(context.Background())
				if err != nil {
					return err
				}

				if len(playback.Queue) == 0 {
					fmt.Println("The queue is empty.")
					return nil
				}

				if count > len(playback.Queue) {
					count = len(playback.Queue)
				}

				for i := 0; i < count; i++ {
					fmt.Printf("%d. %s\n", i+1, spotify.FormatTrack(playback.Queue[i].Track))
				}

				return nil
			},
		},
		{
			name:        "new-device",
			description: "Opens a spotify tab on your default browser.",
			run: func(args []string, _ spotify.Services) error {
				if err := requireNoArgs("new-device", args); err != nil {
					return err
				}

				return browser.OpenURL("http://open.spotify.com")
			},
		},
	}
}

func RunCLI(args []string) error {
	if len(args) == 0 {
		return nil
	}

	cmd, ok := findCommand(args[0])
	if !ok {
		return fmt.Errorf("vibel unknown command %q, try `vibel help`", args[0])
	}

	return cmd.execute(args[1:])
}

func findCommand(name string) (command, bool) {
	for _, cmd := range commandList() {
		if cmd.matches(name) {
			return cmd, true
		}
	}

	return command{}, false
}

func printHelp() {
	fmt.Println(`
VIBEL
In order to launch the TUI, just enter "vibel" with no additional flags.

USAGE: vibel [COMMAND] [AMOUNT (if needed)]

Note: The CLI is meant to be used to interface with Spotify for simple actions.
Commands cannot be chained.
For a more in-depth experience, use the TUI.

COMMANDS:`)

	for _, cmd := range commandList() {
		if cmd.hideFromHelp {
			continue
		}

		fmt.Printf("\t\t%s: \n\t\t\t%s\n", cmd.name, cmd.description)
	}
}

func runHelp(args []string, _ spotify.Services) error {
	if err := requireNoArgs("help", args); err != nil {
		return err
	}
	printHelp()
	return nil
}

func requireNoArgs(name string, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: vibel %s", name)
	}

	return nil
}

func requirePositiveCountArg(args []string) (int, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("usage: vibel list <count>")
	}

	count, err := strconv.Atoi(args[0])
	if err != nil || count < 1 {
		return 0, fmt.Errorf("list count must be a positive integer")
	}

	return count, nil
}
