package yca_google

import (
	"net/http"

	yca_google_maps "github.com/yca-software/yca-go-core/google/maps"
	yca_google_oauth "github.com/yca-software/yca-go-core/google/oauth"
	yca_logger "github.com/yca-software/yca-go-core/logger"
)

type Config struct {
	OAuth      *yca_google_oauth.OAuthConfig
	Maps       *yca_google_maps.MapsConfig
	HTTPClient *http.Client
	Logger     yca_logger.Logger
}

// Module groups enabled Google integrations. Nil fields mean that capability was not configured.
type Module struct {
	OAuth yca_google_oauth.OAuth
	Maps  yca_google_maps.Maps
}

// New builds a Module from config. Only enabled capabilities are wired; others remain nil.
func New(cfg Config) *Module {
	m := &Module{}

	if cfg.OAuth != nil && cfg.OAuth.Enabled() {
		m.OAuth = yca_google_oauth.NewOAuthService(*cfg.OAuth, cfg.HTTPClient)
	}
	if cfg.Maps != nil && cfg.Maps.Enabled() {
		m.Maps = yca_google_maps.NewMapsService(*cfg.Maps, cfg.HTTPClient, cfg.Logger)
	}

	return m
}
