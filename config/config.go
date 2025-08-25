package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	ClientID    string
	RedirectURI string
	Scopes      []string
	Port        string
}

func FromEnv() (Config, error) {
	cfg := Config{
		ClientID:    strings.TrimSpace(os.Getenv("SPOTIFY_CLIENT_ID")),
		RedirectURI: strings.TrimSpace(os.Getenv("SPOTIFY_REDIRECT_URI")),
	}
	cfg.Port = "8888"
	scopesRaw := strings.TrimSpace(os.Getenv("SPOTIFY_SCOPES"))
	if scopesRaw != "" {
		cfg.Scopes = strings.Fields(scopesRaw)
	}
	if cfg.ClientID == "" || cfg.RedirectURI == "" {
		return Config{}, errors.New("missing ClientID or RedirectID")
	}
	return cfg, nil
}
