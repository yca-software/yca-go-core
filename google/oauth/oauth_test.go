package yca_google_oauth_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	yca_google_oauth "github.com/yca-software/yca-go-core/google/oauth"
)

type OAuthSuite struct {
	suite.Suite
}

func TestOAuthSuite(t *testing.T) {
	suite.Run(t, new(OAuthSuite))
}

func (s *OAuthSuite) TestOAuthConfig_Enabled() {
	s.False((yca_google_oauth.OAuthConfig{}).Enabled())
	s.False((yca_google_oauth.OAuthConfig{ClientID: "id"}).Enabled())
	s.True((yca_google_oauth.OAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost/callback",
	}).Enabled())
}

func (s *OAuthSuite) TestGetUserInfo_success() {
	client := &http.Client{Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
		rec := httptestResponseRecorder()
		if strings.Contains(req.URL.String(), "/token") {
			_ = json.NewEncoder(rec).Encode(map[string]string{
				"access_token": "access-token",
				"token_type":   "Bearer",
			})
			return rec.Result(), nil
		}
		_ = json.NewEncoder(rec).Encode(yca_google_oauth.UserInfo{
			ID:            "google-user-1",
			Email:         "ada@example.com",
			VerifiedEmail: true,
			Name:          "Ada Lovelace",
		})
		return rec.Result(), nil
	})}

	svc := yca_google_oauth.NewOAuthService(yca_google_oauth.OAuthConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
	}, client)

	info, err := svc.GetUserInfo(context.Background(), "auth-code")
	s.Require().NoError(err)
	s.Equal("google-user-1", info.ID)
	s.Equal("ada@example.com", info.Email)
	s.True(info.VerifiedEmail)
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func httptestResponseRecorder() *responseRecorder {
	return &responseRecorder{header: make(http.Header)}
}

type responseRecorder struct {
	code   int
	header http.Header
	body   strings.Builder
}

func (r *responseRecorder) Header() http.Header { return r.header }

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.code == 0 {
		r.code = http.StatusOK
	}
	return r.body.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) { r.code = statusCode }

func (r *responseRecorder) Result() *http.Response {
	return &http.Response{
		StatusCode: r.code,
		Header:     r.header,
		Body:       io.NopCloser(strings.NewReader(r.body.String())),
	}
}
