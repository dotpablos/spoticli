package spotify

import (
	"errors"

	"github.com/dotpablos/vibel/internal/spotifysession"
	"github.com/zalando/go-keyring"
)

type Services struct {
	Auth     *AuthService
	Playback *PlaybackService
}

func NewServices() Services {
	return Services{
		Auth:     NewAuthService(),
		Playback: NewPlaybackService(),
	}
}

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) HasSavedSession() bool {
	return spotifysession.HasValidSavedToken()
}

func (s *AuthService) BeginLogin() (*LoginSession, error) {
	authenticator, err := spotifysession.NewAuthenticator()
	if err != nil {
		return nil, err
	}

	session, err := spotifysession.StartLogin(authenticator)
	if err != nil {
		return nil, err
	}

	return &LoginSession{
		URL:     session.URL,
		session: session,
	}, nil
}

func (s *AuthService) Login(output bool) error {
	authenticator, err := spotifysession.NewAuthenticator()
	if err != nil {
		return err
	}

	return spotifysession.Login(authenticator, output)
}

func (s *AuthService) Logout() error {
	err := spotifysession.DeleteToken()
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}

	return err
}

type LoginSession struct {
	URL     string
	session *spotifysession.LoginSession
}

func (s *LoginSession) Open() error {
	if s == nil || s.session == nil {
		return nil
	}

	return s.session.Open()
}

func (s *LoginSession) Wait() error {
	if s == nil || s.session == nil {
		return nil
	}

	return s.session.Wait()
}

func (s *LoginSession) Close() {
	if s == nil || s.session == nil {
		return
	}

	s.session.Close()
}
