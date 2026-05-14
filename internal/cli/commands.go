package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dotpablos/vibel/internal/spotifysession"
	"github.com/pkg/browser"
	"github.com/zalando/go-keyring"
	spotifyapi "github.com/zmb3/spotify/v2"
)

type command struct {
	name         string
	aliases      []string
	description  string
	hideFromHelp bool
	requiresAuth bool
	run          func(args []string, client *spotifyapi.Client) error
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
	var client *spotifyapi.Client

	if cmd.requiresAuth {
		var err error
		client, err = spotifysession.NewSpotifyClient(context.Background())
		if err != nil {
			return err
		}
	}

	if err := cmd.run(args, client); err != nil {
		return fmt.Errorf("%s: %w", cmd.name, err)
	}

	return nil
}

func commandList() []command {
	return []command{
		{
			name:         "help",
			aliases:      []string{"-h", "--help"},
			description:  "Show available commands",
			requiresAuth: false,
			run:          runHelp,
		},
		{
			name:         "login",
			description:  "Authenticate with Spotify",
			requiresAuth: false,
			run: func(args []string, _ *spotifyapi.Client) error {
				if err := requireNoArgs("login", args); err != nil {
					return err
				}

				auth, err := spotifysession.NewAuthenticator()
				if err != nil {
					return err
				}

				return spotifysession.Login(auth)
			},
		},
		{
			name:         "logout",
			description:  "Clear the current Spotify session",
			requiresAuth: false,
			run: func(args []string, _ *spotifyapi.Client) error {
				if err := requireNoArgs("logout", args); err != nil {
					return err
				}

				err := spotifysession.DeleteToken()
				if errors.Is(err, keyring.ErrNotFound) {
					return nil
				}

				return err
			},
		},
		{
			name:         "pause",
			description:  "Pause the current Spotify track",
			requiresAuth: true,
			run: func(args []string, client *spotifyapi.Client) error {
				if err := requireNoArgs("pause", args); err != nil {
					return err
				}

				return runPlayerCommand(context.Background(), client, client.Pause)
			},
		},
		{
			name:         "play",
			description:  "Resume the current Spotify track",
			requiresAuth: true,
			run: func(args []string, client *spotifyapi.Client) error {
				if err := requireNoArgs("play", args); err != nil {
					return err
				}

				return runPlayerCommand(context.Background(), client, client.Play)
			},
		},
		{
			name:         "forward",
			description:  "Skip to the next track in the queue",
			requiresAuth: true,
			run: func(args []string, client *spotifyapi.Client) error {
				if err := requireNoArgs("forward", args); err != nil {
					return err
				}

				return runPlayerCommand(context.Background(), client, client.Next)
			},
		},
		{
			name:         "back",
			description:  "Return to the previous track",
			requiresAuth: true,
			run: func(args []string, client *spotifyapi.Client) error {
				if err := requireNoArgs("back", args); err != nil {
					return err
				}

				return runPlayerCommand(context.Background(), client, client.Previous)
			},
		},
		{
			name:         "song",
			description:  "Show the current song name and playback time",
			requiresAuth: true,
			run: func(args []string, client *spotifyapi.Client) error {
				if err := requireNoArgs("song", args); err != nil {
					return err
				}

				current, err := client.PlayerCurrentlyPlaying(context.Background())
				if err != nil {
					return err
				}

				if current == nil || current.Item == nil {
					fmt.Println("Nothing is currently playing.")
					return nil
				}

				fmt.Printf(
					"%s [%s / %s]\n",
					spotifysession.FormatTrack(*current.Item),
					spotifysession.FormatDuration(int(current.Progress)),
					spotifysession.FormatDuration(int(current.Item.Duration)),
				)
				return nil
			},
		},
		{
			name:         "list",
			description:  "List the next <count> songs in the queue",
			requiresAuth: true,
			run: func(args []string, client *spotifyapi.Client) error {
				count, err := requirePositiveCountArg(args)
				if err != nil {
					return err
				}

				queue, err := client.GetQueue(context.Background())
				if err != nil {
					return err
				}

				if len(queue.Items) == 0 {
					fmt.Println("The queue is empty.")
					return nil
				}

				if count > len(queue.Items) {
					count = len(queue.Items)
				}

				for i := 0; i < count; i++ {
					fmt.Printf("%d. %s\n", i+1, spotifysession.FormatTrack(queue.Items[i]))
				}

				return nil
			},
		},
		{
			name:         "new-device",
			description:  "Opens a spotify tab on your default browser.",
			requiresAuth: true,
			run: func(args []string, client *spotifyapi.Client) error {
				browser.OpenURL("http://open.spotify.com")
				return nil
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

func runHelp(args []string, _ *spotifyapi.Client) error {
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

func runPlayerCommand(ctx context.Context, client *spotifyapi.Client, run func(context.Context) error) error {
	if err := run(ctx); err != nil {
		return explainPlayerCommandError(ctx, client, err)
	}

	return nil
}

func explainPlayerCommandError(ctx context.Context, client *spotifyapi.Client, err error) error {
	var spotifyErr spotifyapi.Error
	if !errors.As(err, &spotifyErr) {
		return err
	}

	if client == nil {
		return err
	}

	devices, devicesErr := client.PlayerDevices(ctx)
	if devicesErr != nil {
		return fmt.Errorf("%s. Spotify usually returns this when there is no active controllable device, or the current account/device cannot be controlled over the Web API", spotifyErr.Message)
	}

	return fmt.Errorf("%s", buildRestrictionExplanation(spotifyErr, devices))
}

func buildRestrictionExplanation(spotifyErr spotifyapi.Error, devices []spotifyapi.PlayerDevice) string {
	activeDevices := make([]string, 0, len(devices))
	availableDevices := make([]string, 0, len(devices))
	restrictedDevices := make([]string, 0, len(devices))

	for _, device := range devices {
		label := deviceLabel(device)
		if device.Active {
			activeDevices = append(activeDevices, label)
		}
		if device.Restricted {
			restrictedDevices = append(restrictedDevices, label)
			continue
		}
		availableDevices = append(availableDevices, label)
	}

	if len(devices) == 0 {
		return fmt.Sprintf("%s. No Spotify devices are available for this account. Open Spotify on a phone, desktop app, or web player and start playback once, then retry", spotifyErr.Message)
	}

	if len(activeDevices) == 0 && len(availableDevices) > 0 {
		return fmt.Sprintf("%s. Spotify can see devices (%s), but none are active. Start playback on one of them first, then retry", spotifyErr.Message, strings.Join(availableDevices, ", "))
	}

	if len(activeDevices) > 0 && len(availableDevices) == 0 {
		return fmt.Sprintf("%s. The active device is restricted (%s), so Spotify will not accept Web API player commands for it", spotifyErr.Message, strings.Join(activeDevices, ", "))
	}

	if len(restrictedDevices) > 0 {
		return fmt.Sprintf("%s. Active/available devices: %s. Restricted devices: %s. If playback is still blocked, check that this is a Premium account and that the target device supports Spotify Connect control", spotifyErr.Message, strings.Join(availableDevices, ", "), strings.Join(restrictedDevices, ", "))
	}

	return fmt.Sprintf("%s. Available devices: %s. If playback is still blocked, check that this is a Premium account and that the target device supports Spotify Connect control", spotifyErr.Message, strings.Join(availableDevices, ", "))
}

func deviceLabel(device spotifyapi.PlayerDevice) string {
	label := device.Name
	if label == "" {
		label = "Unnamed device"
	}
	if device.Type != "" {
		label = fmt.Sprintf("%s (%s)", label, device.Type)
	}
	return label
}
