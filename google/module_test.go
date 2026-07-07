package chi_google_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	chi_google "github.com/yca-software/yca-go-core/google"
	chi_google_maps "github.com/yca-software/yca-go-core/google/maps"
	chi_google_oauth "github.com/yca-software/yca-go-core/google/oauth"
)

type ModuleSuite struct {
	suite.Suite
}

func TestModuleSuite(t *testing.T) {
	suite.Run(t, new(ModuleSuite))
}

func (s *ModuleSuite) TestNew_disabledCapabilities() {
	mod := chi_google.New(chi_google.Config{
		OAuth: &chi_google_oauth.OAuthConfig{},
		Maps:  &chi_google_maps.MapsConfig{},
	})
	s.Nil(mod.OAuth)
	s.Nil(mod.Maps)
}

func (s *ModuleSuite) TestNew_oauthOnly() {
	mod := chi_google.New(chi_google.Config{
		OAuth: &chi_google_oauth.OAuthConfig{
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
	mod := chi_google.New(chi_google.Config{
		Maps:       &chi_google_maps.MapsConfig{APIKey: "maps-key"},
		HTTPClient: http.DefaultClient,
	})
	s.Nil(mod.OAuth)
	s.NotNil(mod.Maps)
}

func (s *ModuleSuite) TestNew_bothEnabled() {
	mod := chi_google.New(chi_google.Config{
		OAuth: &chi_google_oauth.OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
		},
		Maps:       &chi_google_maps.MapsConfig{APIKey: "maps-key"},
		HTTPClient: http.DefaultClient,
	})
	s.NotNil(mod.OAuth)
	s.NotNil(mod.Maps)
}
