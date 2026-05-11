package spotifysession

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

func TestRedirectURIPrefersSPOTIFY_REDIRECT_URI(t *testing.T) {
	t.Setenv("SPOTIFY_REDIRECT_URI", "http://127.0.0.1:9000/uri")
	t.Setenv("SPOTIFY_REDIRECT_URL", "http://127.0.0.1:8080/url")

	if got := redirectURI(); got != "http://127.0.0.1:9000/uri" {
		t.Fatalf("redirectURI() = %q, want %q", got, "http://127.0.0.1:9000/uri")
	}
}

func TestRedirectURIFallsBackToSPOTIFY_REDIRECT_URL(t *testing.T) {
	t.Setenv("SPOTIFY_REDIRECT_URI", "")
	t.Setenv("SPOTIFY_REDIRECT_URL", "http://127.0.0.1:8080/url")

	if got := redirectURI(); got != "http://127.0.0.1:8080/url" {
		t.Fatalf("redirectURI() = %q, want %q", got, "http://127.0.0.1:8080/url")
	}
}

func TestRedirectURIFallsBackToDefault(t *testing.T) {
	t.Setenv("SPOTIFY_REDIRECT_URI", "")
	t.Setenv("SPOTIFY_REDIRECT_URL", "")

	if got := redirectURI(); got != defaultRedirectURI {
		t.Fatalf("redirectURI() = %q, want %q", got, defaultRedirectURI)
	}
}

func TestMarshalUnmarshalTokenRoundTrip(t *testing.T) {
	token := &oauth2.Token{
		AccessToken:  "access-token",
		TokenType:    "Bearer",
		RefreshToken: "refresh-token",
		Expiry:       time.Now().UTC().Add(30 * time.Minute).Round(0),
	}

	raw, err := marshalToken(token)
	if err != nil {
		t.Fatalf("marshalToken() error = %v", err)
	}

	got, err := unmarshalToken(raw)
	if err != nil {
		t.Fatalf("unmarshalToken() error = %v", err)
	}

	if got.AccessToken != token.AccessToken {
		t.Fatalf("AccessToken = %q, want %q", got.AccessToken, token.AccessToken)
	}
	if got.RefreshToken != token.RefreshToken {
		t.Fatalf("RefreshToken = %q, want %q", got.RefreshToken, token.RefreshToken)
	}
	if !got.Expiry.Equal(token.Expiry) {
		t.Fatalf("Expiry = %v, want %v", got.Expiry, token.Expiry)
	}
}

func TestUnmarshalTokenSupportsLegacyAccessToken(t *testing.T) {
	got, err := unmarshalToken("legacy-access-token")
	if err != nil {
		t.Fatalf("unmarshalToken() error = %v", err)
	}

	if got.AccessToken != "legacy-access-token" {
		t.Fatalf("AccessToken = %q, want %q", got.AccessToken, "legacy-access-token")
	}
	if got.TokenType != "Bearer" {
		t.Fatalf("TokenType = %q, want %q", got.TokenType, "Bearer")
	}
}

func TestHandleLoginCallbackRejectsStateMismatch(t *testing.T) {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL("http://127.0.0.1:8080/callback"),
		spotifyauth.WithClientID("client-id"),
		spotifyauth.WithClientSecret("client-secret"),
	)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=wrong&code=test-code", nil)
	rec := httptest.NewRecorder()

	err := handleLoginCallback(
		rec,
		req,
		"expected",
		func(ctx context.Context, state string, r *http.Request) (*oauth2.Token, error) {
			return auth.Token(ctx, state, r)
		},
		func(*oauth2.Token) error { return nil },
	)
	if err == nil {
		t.Fatal("handleLoginCallback() error = nil, want non-nil")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(err.Error(), "state parameter doesn't match") {
		t.Fatalf("error = %q, want state mismatch", err)
	}
}

func TestHandleLoginCallbackRejectsMissingCode(t *testing.T) {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL("http://127.0.0.1:8080/callback"),
		spotifyauth.WithClientID("client-id"),
		spotifyauth.WithClientSecret("client-secret"),
	)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=expected", nil)
	rec := httptest.NewRecorder()

	err := handleLoginCallback(
		rec,
		req,
		"expected",
		func(ctx context.Context, state string, r *http.Request) (*oauth2.Token, error) {
			return auth.Token(ctx, state, r)
		},
		func(*oauth2.Token) error { return nil },
	)
	if err == nil {
		t.Fatal("handleLoginCallback() error = nil, want non-nil")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(err.Error(), "didn't get access code") {
		t.Fatalf("error = %q, want missing code", err)
	}
}

func TestHandleLoginCallbackSuccess(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/callback?state=expected&code=test-code", nil)
	rec := httptest.NewRecorder()

	token := &oauth2.Token{AccessToken: "access-token", TokenType: "Bearer"}
	saved := false

	err := handleLoginCallback(
		rec,
		req,
		"expected",
		func(context.Context, string, *http.Request) (*oauth2.Token, error) {
			return token, nil
		},
		func(got *oauth2.Token) error {
			saved = true
			if got != token {
				t.Fatalf("save token pointer mismatch")
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("handleLoginCallback() error = %v", err)
	}
	if !saved {
		t.Fatal("save() was not called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want %q", got, "text/html; charset=utf-8")
	}
	if !strings.Contains(rec.Body.String(), "Authentication complete.") {
		t.Fatalf("body missing success page content: %q", rec.Body.String())
	}
}

func TestHandleLoginCallbackSaveFailure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/callback?state=expected&code=test-code", nil)
	rec := httptest.NewRecorder()

	saveErr := errors.New("keyring unavailable")
	err := handleLoginCallback(
		rec,
		req,
		"expected",
		func(context.Context, string, *http.Request) (*oauth2.Token, error) {
			return &oauth2.Token{AccessToken: "access-token", TokenType: "Bearer"}, nil
		},
		func(*oauth2.Token) error {
			return saveErr
		},
	)
	if err == nil {
		t.Fatal("handleLoginCallback() error = nil, want non-nil")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(err.Error(), saveErr.Error()) {
		t.Fatalf("error = %q, want %q", err, saveErr)
	}
}
