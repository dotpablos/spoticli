package player

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dotpablos/vibel/internal/app"
	"github.com/dotpablos/vibel/internal/spotify"
	"github.com/dotpablos/vibel/internal/ui"
)

type controlButton struct {
	icon   string
	label  string
	accent lipgloss.Color
}

type Model struct {
	app.ScreenLayout
	service  *spotify.PlaybackService
	commands []app.Command
	playback spotify.PlaybackSnapshot
	status   string
	statusOK bool
	loaded   bool
}

type playerCommandResultMsg struct {
	status    string
	err       error
	isPlaying *bool
}

type playerPlaybackDataMsg struct {
	playback spotify.PlaybackSnapshot
	err      error
}

type fetchPlaybackMsg struct{}

type queueEntry struct {
	Position  int
	Track     spotify.Track
	IsCurrent bool
}

var (
	demoTrackNeonGravity = spotify.Track{
		Title:      "Neon Gravity",
		Album:      "Blood Machines OST",
		Artist:     "Carpenter Brut",
		DurationMs: 4*60*1000 + 7*1000,
	}
	demoTrackNightcall = spotify.Track{
		Title:      "Nightcall",
		Album:      "OutRun",
		Artist:     "Kavinsky",
		DurationMs: 4*60*1000 + 18*1000,
	}
	demoPlayback = spotify.PlaybackSnapshot{
		Current:      demoTrackNeonGravity,
		ProgressMs:   92 * 1000,
		LastSyncedAt: time.Unix(0, 0),
		IsPlaying:    true,
	}
	demoQueue = []spotify.QueueTrack{
		{
			Position: 1,
			Track: spotify.Track{
				Title:      "Turbo Killer",
				Album:      "Trilogy",
				Artist:     "Carpenter Brut",
				DurationMs: 5*60*1000 + 52*1000,
			},
		},
		{
			Position: 2,
			Track: spotify.Track{
				Title:      "Hang 'Em All",
				Album:      "Trilogy",
				Artist:     "Carpenter Brut",
				DurationMs: 3*60*1000 + 44*1000,
			},
		},
		{
			Position: 3,
			Track: spotify.Track{
				Title:      "Le Perv",
				Album:      "Trilogy",
				Artist:     "Carpenter Brut",
				DurationMs: 4*60*1000 + 29*1000,
			},
		},
		{
			Position: 4,
			Track: spotify.Track{
				Title:      "Friday Night",
				Album:      "OutRun",
				Artist:     "Kavinsky",
				DurationMs: 4*60*1000 + 12*1000,
			},
		},
		{
			Position: 5,
			Track:    demoTrackNightcall,
		},
		{
			Position: 6,
			Track: spotify.Track{
				Title:      "ProtoVision",
				Album:      "OutRun",
				Artist:     "Kavinsky",
				DurationMs: 5*60*1000 + 1*1000,
			},
		},
		{
			Position: 7,
			Track: spotify.Track{
				Title:      "Odd Look",
				Album:      "Odd Look",
				Artist:     "Kavinsky ft. The Weeknd",
				DurationMs: 3*60*1000 + 58*1000,
			},
		},
	}
)

func New(service *spotify.PlaybackService) *Model {
	if service == nil {
		service = spotify.NewPlaybackService()
	}

	return &Model{service: service}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.ensureCommands()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Resize(msg)
		if !m.loaded {
			return m.fetchPlaybackInfoCmd()
		}
	case app.NavigateMsg:
		if msg.Screen == app.ScreenPlayer {
			return m.fetchPlaybackInfoCmd()
		}
	case fetchPlaybackMsg:
		return m.fetchPlaybackInfoCmd()
	case playerPlaybackDataMsg:
		if msg.err != nil {
			m.setStatus(msg.err.Error(), false)
			m.loaded = true
			return pollPlaybackCmd()
		}

		m.applyPlayback(msg.playback)
		return pollPlaybackCmd()
	case playerCommandResultMsg:
		if msg.err != nil {
			m.setStatus(msg.err.Error(), false)
			return nil
		}

		if msg.isPlaying != nil {
			m.setPlaybackState(*msg.isPlaying)
		}

		m.setStatus(msg.status, true)
		return m.fetchPlaybackInfoCmd()
	case tea.KeyMsg:
		return app.CommandForKey(msg.String(), m.commands)
	}

	return nil
}

func (m *Model) View() string {
	m.ensureCommands()

	width := m.ContentWidth()
	height := m.ContentHeight()
	gap := 1
	commandBar := app.RenderCommandBar(m.commands, width-4, "╰", "╯", "─")
	panelHeight := max(height-1, 1)

	leftWidth := max((width-gap)/3, 20)
	queueWidth := max(width-gap-leftWidth, 24)

	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderNowPlayingPanel(leftWidth, panelHeight),
		" ",
		m.renderQueuePanel(queueWidth, panelHeight),
	)

	content := lipgloss.JoinVertical(lipgloss.Left, panels, commandBar)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Left,
		lipgloss.Top,
		lipgloss.NewStyle().
			Foreground(ui.Text).
			Padding(0, 1).
			Render(content),
	)
}
