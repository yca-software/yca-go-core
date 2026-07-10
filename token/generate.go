package yca_token

import (
	"crypto/rand"
	"encoding/base64"
)

const (
	TokenLength = 32
)

// GenerateOpaqueToken returns a URL-safe opaque token without base64 padding.
// Use for refresh tokens, invites, and API keys in application starters.
func GenerateOpaqueToken() (string, error) {
	tokenBytes := make([]byte, TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(tokenBytes), nil
}
