package spotify

import (
	"context"
	"strings"
	"time"

	"github.com/dotpablos/vibel/internal/spotifysession"
	spotifyapi "github.com/zmb3/spotify/v2"
)

type Track struct {
	Title      string
	Album      string
	Artist     string
	DurationMs int
}

type QueueTrack struct {
	Position int
	Track    Track
}

type PlaybackSnapshot struct {
	Current      Track
	Queue        []QueueTrack
	ProgressMs   int
	LastSyncedAt time.Time
	IsPlaying    bool
}

func (s PlaybackSnapshot) HasCurrentTrack() bool {
	return s.Current.Title != ""
}

type PlaybackService struct{}

func NewPlaybackService() *PlaybackService {
	return &PlaybackService{}
}

func (s *PlaybackService) Snapshot(ctx context.Context) (PlaybackSnapshot, error) {
	client, err := spotifysession.NewSpotifyClient(ctx)
	if err != nil {
		return PlaybackSnapshot{}, err
	}

	current, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return PlaybackSnapshot{}, err
	}

	queue, err := client.GetQueue(ctx)
	if err != nil {
		return PlaybackSnapshot{}, err
	}

	return buildPlaybackSnapshot(current, queue), nil
}

func (s *PlaybackService) Toggle(ctx context.Context) (bool, error) {
	client, err := spotifysession.NewSpotifyClient(ctx)
	if err != nil {
		return false, err
	}

	current, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return false, err
	}

	nextState := true
	run := client.Play
	if current != nil && current.Playing {
		nextState = false
		run = client.Pause
	}

	if err := spotifysession.RunPlayerCommand(ctx, client, run); err != nil {
		return false, err
	}

	return nextState, nil
}

func (s *PlaybackService) Play(ctx context.Context) error {
	return spotifysession.ExecutePlayerCommand(ctx, func(ctx context.Context, client *spotifyapi.Client) error {
		return client.Play(ctx)
	})
}

func (s *PlaybackService) Pause(ctx context.Context) error {
	return spotifysession.ExecutePlayerCommand(ctx, func(ctx context.Context, client *spotifyapi.Client) error {
		return client.Pause(ctx)
	})
}

func (s *PlaybackService) Next(ctx context.Context) error {
	return spotifysession.ExecutePlayerCommand(ctx, func(ctx context.Context, client *spotifyapi.Client) error {
		return client.Next(ctx)
	})
}

func (s *PlaybackService) Previous(ctx context.Context) error {
	return spotifysession.ExecutePlayerCommand(ctx, func(ctx context.Context, client *spotifyapi.Client) error {
		return client.Previous(ctx)
	})
}

func buildPlaybackSnapshot(current *spotifyapi.CurrentlyPlaying, queue *spotifyapi.Queue) PlaybackSnapshot {
	snapshot := PlaybackSnapshot{
		LastSyncedAt: time.Now(),
		IsPlaying:    current != nil && current.Playing,
	}

	if current != nil && current.Item != nil {
		snapshot.Current = trackFromFullTrack(*current.Item)
		snapshot.ProgressMs = int(current.Progress)
	} else if queue != nil && queue.CurrentlyPlaying.Name != "" {
		snapshot.Current = trackFromFullTrack(queue.CurrentlyPlaying)
	}

	items := queueItems(queue)
	snapshot.Queue = make([]QueueTrack, 0, len(items))
	for index, item := range items {
		snapshot.Queue = append(snapshot.Queue, QueueTrack{
			Position: index + 1,
			Track:    trackFromFullTrack(item),
		})
	}

	return snapshot
}

func queueItems(queue *spotifyapi.Queue) []spotifyapi.FullTrack {
	if queue == nil {
		return nil
	}

	return queue.Items
}

func trackFromFullTrack(track spotifyapi.FullTrack) Track {
	return Track{
		Title:      track.Name,
		Album:      track.Album.Name,
		Artist:     joinArtistNames(track.Artists),
		DurationMs: int(track.Duration),
	}
}

func joinArtistNames(artists []spotifyapi.SimpleArtist) string {
	if len(artists) == 0 {
		return ""
	}

	names := make([]string, 0, len(artists))
	for _, artist := range artists {
		if artist.Name != "" {
			names = append(names, artist.Name)
		}
	}

	return strings.Join(names, ", ")
}
