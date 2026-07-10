package yca_google_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	yca_google "github.com/yca-software/yca-go-core/google"
	yca_google_maps "github.com/yca-software/yca-go-core/google/maps"
	yca_google_oauth "github.com/yca-software/yca-go-core/google/oauth"
)

type ModuleSuite struct {
	suite.Suite
}

func TestModuleSuite(t *testing.T) {
	suite.Run(t, new(ModuleSuite))
}

func (s *ModuleSuite) TestNew_disabledCapabilities() {
	mod := yca_google.New(yca_google.Config{
		OAuth: &yca_google_oauth.OAuthConfig{},
		Maps:  &yca_google_maps.MapsConfig{},
	})
	s.Nil(mod.OAuth)
	s.Nil(mod.Maps)
}

func (s *ModuleSuite) TestNew_oauthOnly() {
	mod := yca_google.New(yca_google.Config{
		OAuth: &yca_google_oauth.OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
		},
		HTTPClient: http.DefaultClient,
	})
	s.NotNil(mod.OAuth)
	s.Nil(mod.Maps)
}

func (s *ModuleSuite) TestNew_mapsOnly() {
	mod := yca_google.New(yca_google.Config{
		Maps:       &yca_google_maps.MapsConfig{APIKey: "maps-key"},
		HTTPClient: http.DefaultClient,
	})
	s.Nil(mod.OAuth)
	s.NotNil(mod.Maps)
}

func (s *ModuleSuite) TestNew_bothEnabled() {
	mod := yca_google.New(yca_google.Config{
		OAuth: &yca_google_oauth.OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
		},
		Maps:       &yca_google_maps.MapsConfig{APIKey: "maps-key"},
		HTTPClient: http.DefaultClient,
	})
	s.NotNil(mod.OAuth)
	s.NotNil(mod.Maps)
}
