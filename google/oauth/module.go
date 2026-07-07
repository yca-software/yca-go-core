package chi_google_oauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthConfig enables Google OAuth (userinfo).
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func (c OAuthConfig) Enabled() bool {
	return c.ClientID != "" && c.ClientSecret != "" && c.RedirectURL != ""
}

// OAuth provides Google sign-in user profile lookup.
type OAuth interface {
	GetUserInfo(ctx context.Context, code string) (*UserInfo, error)
}

type oauthService struct {
	oauthConfig *oauth2.Config
	httpClient  *http.Client
}

func NewOAuthService(cfg OAuthConfig, httpClient *http.Client) OAuth {
	return &oauthService{
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		httpClient: httpClient,
	}
}

func (s *oauthService) GetUserInfo(ctx context.Context, code string) (*UserInfo, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, s.httpClient)

	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info UserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}

	return &info, nil
}
