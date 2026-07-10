package yca_token

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// Hasher hashes opaque tokens for storage and lookup using HMAC-SHA256.
type Hasher struct {
	pepper string
}

// NewHasher returns a token hasher. Pass pepper from app config (e.g. TOKEN_HASH_PEPPER).
func NewHasher(pepper string) *Hasher {
	if pepper == "" {
		panic("token: pepper is required")
	}
	return &Hasher{pepper: pepper}
}

// Hash returns the HMAC-SHA256 hex digest for storage and DB lookup.
func (h *Hasher) Hash(token string) string {
	mac := hmac.New(sha256.New, []byte(h.pepper))
	_, _ = mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify reports whether storedHash matches the plaintext token.
func (h *Hasher) Verify(storedHash, token string) bool {
	candidate := h.Hash(token)
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(candidate)) == 1
}
