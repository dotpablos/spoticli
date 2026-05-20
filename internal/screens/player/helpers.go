package player

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dotpablos/vibel/internal/app"
	"github.com/dotpablos/vibel/internal/spotify"
	"github.com/dotpablos/vibel/internal/ui"
)

func pollPlaybackCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return fetchPlaybackMsg{}
	})
}

func (m *Model) ensureCommands() {
	if len(m.commands) == 0 {
		m.commands = m.updateCommands()
	}
}

func (m *Model) setStatus(status string, ok bool) {
	m.status = status
	m.statusOK = ok
}

func (m *Model) setPlaybackState(isPlaying bool) {
	m.playback.IsPlaying = isPlaying
	m.loaded = true
	m.commands = m.updateCommands()
}

func (m *Model) applyPlayback(playback spotify.PlaybackSnapshot) {
	m.playback = playback
	m.loaded = true
	m.commands = m.updateCommands()
}

func (m *Model) queueEntries() []queueEntry {
	entries := make([]queueEntry, 0, len(m.playback.Queue)+1)
	if m.playback.HasCurrentTrack() {
		entries = append(entries, queueEntry{
			Track:     m.playback.Current,
			IsCurrent: true,
		})
	}

	for _, track := range m.playback.Queue {
		entries = append(entries, queueEntry{
			Position: track.Position,
			Track:    track.Track,
		})
	}

	return entries
}

func demoQueueEntries() []queueEntry {
	entries := []queueEntry{
		{
			Track:     demoPlayback.Current,
			IsCurrent: true,
		},
	}

	for _, track := range demoQueue {
		entries = append(entries, queueEntry{
			Position: track.Position,
			Track:    track.Track,
		})
	}

	return entries
}

func (m *Model) updateCommands() []app.Command {
	commands := append([]app.Command{}, app.GlobalCommands()...)
	toggle := m.toggleButton()

	commands = append(commands,
		app.Command{
			Name: toggle.label,
			Key:  "p",
			Exec: func() tea.Cmd {
				return m.runTogglePlayerCommand()
			},
		},
		app.Command{
			Name: "skip",
			Key:  "n",
			Exec: func() tea.Cmd {
				return m.runPlayerCommand("skipped to next track", func(ctx context.Context) error {
					return m.service.Next(ctx)
				})
			},
		},
	)

	return commands
}

func (m *Model) fetchPlaybackInfoCmd() tea.Cmd {
	return func() tea.Msg {
		playback, err := m.service.Snapshot(context.Background())
		if err != nil {
			return playerPlaybackDataMsg{err: err}
		}

		return playerPlaybackDataMsg{playback: playback}
	}
}

func (m *Model) runTogglePlayerCommand() tea.Cmd {
	return func() tea.Msg {
		isPlaying, err := m.service.Toggle(context.Background())
		if err != nil {
			return playerCommandResultMsg{err: err}
		}

		status := "playback started"
		if !isPlaying {
			status = "playback stopped"
		}

		return playerCommandResultMsg{status: status, isPlaying: &isPlaying}
	}
}

func (m *Model) runPlayerCommand(successStatus string, run func(context.Context) error) tea.Cmd {
	return func() tea.Msg {
		if err := run(context.Background()); err != nil {
			return playerCommandResultMsg{err: err}
		}

		return playerCommandResultMsg{status: successStatus}
	}
}

func (m *Model) toggleButton() controlButton {
	if m.playback.IsPlaying {
		return controlButton{icon: "||", label: "stop", accent: ui.Red}
	}

	return controlButton{icon: ">", label: "play", accent: ui.Green}
}

func progressRatio(progressMs int, durationMs int) float64 {
	if durationMs <= 0 {
		return 0
	}

	return min(1.0, max(float64(progressMs)/float64(durationMs), 0.0))
}
