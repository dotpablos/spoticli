package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	spotifyapi "github.com/zmb3/spotify/v2"
)

func TestExplainPlayerCommandErrorReturnsOriginalNonSpotifyError(t *testing.T) {
	err := errors.New("boom")

	got := explainPlayerCommandError(context.Background(), nil, err)
	if !errors.Is(got, err) {
		t.Fatalf("expected original error, got %v", got)
	}
}

func TestExplainPlayerCommandErrorReturnsOriginalForOtherSpotifyErrors(t *testing.T) {
	err := spotifyapi.Error{Status: 404, Message: "not found"}

	got := explainPlayerCommandError(context.Background(), nil, err)
	if got.Error() != err.Error() {
		t.Fatalf("got %q, want %q", got.Error(), err.Error())
	}
}

func TestDeviceLabel(t *testing.T) {
	got := deviceLabel(spotifyapi.PlayerDevice{Name: "MacBook", Type: "Computer"})
	if got != "MacBook (Computer)" {
		t.Fatalf("deviceLabel() = %q", got)
	}
}

func TestDeviceLabelUnnamed(t *testing.T) {
	got := deviceLabel(spotifyapi.PlayerDevice{})
	if got != "Unnamed device" {
		t.Fatalf("deviceLabel() = %q", got)
	}
}

func TestRestrictionExplanationNoDevicesMessage(t *testing.T) {
	message := buildRestrictionExplanation(
		spotifyapi.Error{Status: 403, Message: "Restriction violated"},
		nil,
	)

	if !strings.Contains(message, "No Spotify devices are available") {
		t.Fatalf("message = %q", message)
	}
}

func TestRestrictionExplanationInactiveDevicesMessage(t *testing.T) {
	message := buildRestrictionExplanation(
		spotifyapi.Error{Status: 403, Message: "Restriction violated"},
		[]spotifyapi.PlayerDevice{
			{Name: "Office", Type: "Computer"},
		},
	)

	if !strings.Contains(message, "none are active") {
		t.Fatalf("message = %q", message)
	}
}

func TestRestrictionExplanationRestrictedActiveDeviceMessage(t *testing.T) {
	message := buildRestrictionExplanation(
		spotifyapi.Error{Status: 403, Message: "Restriction violated"},
		[]spotifyapi.PlayerDevice{
			{Name: "Phone", Type: "Smartphone", Active: true, Restricted: true},
		},
	)

	if !strings.Contains(message, "active device is restricted") {
		t.Fatalf("message = %q", message)
	}
}

func TestRestrictionExplanationGeneralMessage(t *testing.T) {
	message := buildRestrictionExplanation(
		spotifyapi.Error{Status: 403, Message: "Restriction violated"},
		[]spotifyapi.PlayerDevice{
			{Name: "MacBook", Type: "Computer", Active: true},
		},
	)

	if !strings.Contains(message, "Premium account") {
		t.Fatalf("message = %q", message)
	}
}
