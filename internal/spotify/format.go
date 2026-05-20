package spotify

import (
	"fmt"
	"time"
)

func FormatTrack(track Track) string {
	if track.Artist == "" {
		return track.Title
	}

	return fmt.Sprintf("%s - %s", track.Artist, track.Title)
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
