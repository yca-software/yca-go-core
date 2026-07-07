package chi_google

import (
	"net/http"

	chi_google_maps "github.com/yca-software/yca-go-core/google/maps"
	chi_google_oauth "github.com/yca-software/yca-go-core/google/oauth"
	chi_logger "github.com/yca-software/yca-go-core/logger"
)

type Config struct {
	OAuth      *chi_google_oauth.OAuthConfig
	Maps       *chi_google_maps.MapsConfig
	HTTPClient *http.Client
	Logger     chi_logger.Logger
}

// Module groups enabled Google integrations. Nil fields mean that capability was not configured.
type Module struct {
	OAuth chi_google_oauth.OAuth
	Maps  chi_google_maps.Maps
}

// New builds a Module from config. Only enabled capabilities are wired; others remain nil.
func New(cfg Config) *Module {
	m := &Module{}

	if cfg.OAuth != nil && cfg.OAuth.Enabled() {
		m.OAuth = chi_google_oauth.NewOAuthService(*cfg.OAuth, cfg.HTTPClient)
	}
	if cfg.Maps != nil && cfg.Maps.Enabled() {
		m.Maps = chi_google_maps.NewMapsService(*cfg.Maps, cfg.HTTPClient, cfg.Logger)
	}

	return m
}
