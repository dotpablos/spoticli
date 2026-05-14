package spotifysession

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
	spotifyapi "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func NewSpotifyClient(ctx context.Context) (*spotifyapi.Client, error) {
	token, err := LoadToken()
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, fmt.Errorf("no Spotify session found; run `vibel login` first")
		}

		return nil, fmt.Errorf("load Spotify token: %w", err)
	}

	if token.Valid() {
		return spotifyapi.New(oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))), nil
	}

	if token.RefreshToken == "" {
		return nil, fmt.Errorf("stored Spotify session has expired; run `vibel logout` then `vibel login` to refresh it")
	}

	auth, err := NewAuthenticator()
	if err != nil {
		return nil, fmt.Errorf("refresh Spotify session: %w", err)
	}

	return spotifyapi.New(auth.Client(ctx, token)), nil
}

func FormatTrack(track spotifyapi.FullTrack) string {
	artists := make([]string, 0, len(track.Artists))
	for _, artist := range track.Artists {
		artists = append(artists, artist.Name)
	}

	if len(artists) == 0 {
		return track.Name
	}

	return fmt.Sprintf("%s - %s", strings.Join(artists, ", "), track.Name)
}

func FormatDuration(milliseconds int) string {
	duration := time.Duration(milliseconds) * time.Millisecond
	totalSeconds := int(duration / time.Second)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}

	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
