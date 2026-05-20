package spotifysession

import (
	"context"
	"errors"
	"fmt"
	"strings"

	spotifyapi "github.com/zmb3/spotify/v2"
)

func RunPlayerCommand(ctx context.Context, client *spotifyapi.Client, run func(context.Context) error) error {
	if err := run(ctx); err != nil {
		return explainPlayerCommandError(ctx, client, err)
	}

	return nil
}

func ExecutePlayerCommand(ctx context.Context, run func(context.Context, *spotifyapi.Client) error) error {
	client, err := NewSpotifyClient(ctx)
	if err != nil {
		return err
	}

	return RunPlayerCommand(ctx, client, func(ctx context.Context) error {
		return run(ctx, client)
	})
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
		label += " (" + device.Type + ")"
	}

	return label
}
