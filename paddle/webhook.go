package chi_paddle

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

const maxSignatureAge = 5 * time.Minute

// WebhookEvent is the top-level Paddle webhook payload.
type WebhookEvent struct {
	EventID   string         `json:"event_id"`
	EventType string         `json:"event_type"`
	Data      map[string]any `json:"data"`
}

// VerifyWebhook validates a Paddle-Signature header (ts=...;h1=...) against the raw body.
func VerifyWebhook(secret string, body []byte, signatureHeader string) bool {
	if secret == "" || len(body) == 0 {
		return false
	}

	ts, h1, ok := parsePaddleSignature(signatureHeader)
	if !ok || h1 == "" {
		return false
	}

	timestamp, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return false
	}

	eventTime := time.Unix(timestamp, 0)
	if time.Since(eventTime) > maxSignatureAge || eventTime.After(time.Now().Add(time.Minute)) {
		return false
	}

	payload := ts + ":" + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(h1))
}

func parsePaddleSignature(header string) (ts, h1 string, ok bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", "", false
	}
	for _, part := range strings.Split(header, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "ts=") {
			ts = strings.TrimPrefix(part, "ts=")
		}
		if strings.HasPrefix(part, "h1=") {
			h1 = strings.TrimPrefix(part, "h1=")
		}
	}
	return ts, h1, ts != ""
}
