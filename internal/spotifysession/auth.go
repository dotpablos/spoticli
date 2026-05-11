package spotifysession

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/zalando/go-keyring"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	service            = "spoticli"
	user               = "spotify-token"
	defaultRedirectURI = "http://127.0.0.1:8080/callback"
	loginTimeout       = 2 * time.Minute
)

//go:embed callback.html
var callbackPage []byte

type loginResult struct {
	err error
}

func NewAuthenticator() (*spotifyauth.Authenticator, error) {
	if err := loadEnv(); err != nil {
		return nil, err
	}

	uri := redirectURI()
	if _, _, err := listenAddrFromRedirectURI(uri); err != nil {
		return nil, err
	}

	return spotifyauth.New(
		spotifyauth.WithRedirectURL(uri),
		spotifyauth.WithClientID(os.Getenv("SPOTIFY_ID")),
		spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_SECRET")),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserModifyPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
		),
	), nil
}

func Login(auth *spotifyauth.Authenticator) error {
	state, err := randomState()
	if err != nil {
		return err
	}

	addr, callbackPath, err := listenAddrFromRedirectURI(redirectURI())
	if err != nil {
		return err
	}

	results := make(chan loginResult, 1)
	serverErrors := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		err := handleLoginCallback(
			w,
			r,
			state,
			func(ctx context.Context, expectedState string, req *http.Request) (*oauth2.Token, error) {
				return auth.Token(ctx, expectedState, req)
			},
			SaveToken,
		)

		select {
		case results <- loginResult{err: err}:
		default:
		}
	})

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen for Spotify callback on %s: %w", addr, err)
	}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("serve Spotify callback: %w", err)
		}
	}()

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following URL: " + url)
	fmt.Println("Waiting for Spotify to complete authentication...")
	if err := browser.OpenURL(url); err != nil {
		fmt.Printf("Could not open the browser automatically: %v\n", err)
	}

	defer shutdownServer(server)

	select {
	case result := <-results:
		return result.err
	case err := <-serverErrors:
		return err
	case <-time.After(loginTimeout):
		return fmt.Errorf("timed out waiting for Spotify callback after %s", loginTimeout)
	}
}

func SaveToken(token *oauth2.Token) error {
	raw, err := marshalToken(token)
	if err != nil {
		return err
	}

	return keyring.Set(service, user, raw)
}

func LoadToken() (*oauth2.Token, error) {
	raw, err := keyring.Get(service, user)
	if err != nil {
		return nil, err
	}

	return unmarshalToken(raw)
}

func DeleteToken() error {
	return keyring.Delete(service, user)
}

func redirectURI() string {
	if uri := os.Getenv("SPOTIFY_REDIRECT_URI"); uri != "" {
		return uri
	}

	if uri := os.Getenv("SPOTIFY_REDIRECT_URL"); uri != "" {
		return uri
	}

	return defaultRedirectURI
}

func listenAddrFromRedirectURI(uri string) (string, string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", "", fmt.Errorf("parse Spotify redirect URI %q: %w", uri, err)
	}

	if parsed.Scheme != "http" {
		return "", "", fmt.Errorf("Spotify redirect URI must use http, got %q", parsed.Scheme)
	}

	if parsed.Host == "" {
		return "", "", fmt.Errorf("Spotify redirect URI must include a host")
	}

	if _, _, err := net.SplitHostPort(parsed.Host); err != nil {
		return "", "", fmt.Errorf("Spotify redirect URI host must include a port: %w", err)
	}

	path := parsed.EscapedPath()
	if path == "" {
		path = "/callback"
	}

	return parsed.Host, path, nil
}

func successPageHTML() []byte {
	return callbackPage
}

func marshalToken(token *oauth2.Token) (string, error) {
	if token == nil {
		return "", fmt.Errorf("Spotify token is nil")
	}

	raw, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("marshal Spotify token: %w", err)
	}

	return string(raw), nil
}

func unmarshalToken(raw string) (*oauth2.Token, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("stored Spotify token is empty")
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(raw), &token); err == nil {
		if token.AccessToken == "" {
			return nil, fmt.Errorf("stored Spotify token is missing an access token")
		}

		if token.TokenType == "" {
			token.TokenType = "Bearer"
		}

		return &token, nil
	}

	return &oauth2.Token{
		AccessToken: raw,
		TokenType:   "Bearer",
	}, nil
}

func randomState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate Spotify auth state: %w", err)
	}

	return hex.EncodeToString(buf), nil
}

func handleLoginCallback(
	w http.ResponseWriter,
	r *http.Request,
	state string,
	exchange func(context.Context, string, *http.Request) (*oauth2.Token, error),
	save func(*oauth2.Token) error,
) error {
	token, err := exchange(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Spotify authentication failed. Return to your terminal for details.", http.StatusBadRequest)
		return fmt.Errorf("exchange Spotify token: %w", err)
	}

	if err := save(token); err != nil {
		http.Error(w, "Spotify authentication succeeded, but saving the session failed.", http.StatusInternalServerError)
		return fmt.Errorf("save Spotify token: %w", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(successPageHTML()); err != nil {
		return fmt.Errorf("write Spotify callback response: %w", err)
	}

	return nil
}

func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
