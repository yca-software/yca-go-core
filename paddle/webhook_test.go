package chi_paddle

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type WebhookSuite struct {
	suite.Suite
}

func TestWebhookSuite(t *testing.T) {
	suite.Run(t, new(WebhookSuite))
}

func (s *WebhookSuite) sign(secret string, body []byte, ts int64) string {
	payload := strconv.FormatInt(ts, 10) + ":" + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return fmt.Sprintf("ts=%d;h1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

func (s *WebhookSuite) TestVerifyWebhook_valid() {
	secret := "whsec_test"
	body := []byte(`{"event_type":"subscription.created"}`)
	ts := time.Now().Unix()
	header := s.sign(secret, body, ts)

	s.True(VerifyWebhook(secret, body, header))
}

func (s *WebhookSuite) TestVerifyWebhook_invalidSecret() {
	body := []byte(`{"event_type":"subscription.created"}`)
	ts := time.Now().Unix()
	header := s.sign("whsec_real", body, ts)

	s.False(VerifyWebhook("whsec_wrong", body, header))
}

func (s *WebhookSuite) TestVerifyWebhook_tamperedBody() {
	secret := "whsec_test"
	body := []byte(`{"event_type":"subscription.created"}`)
	ts := time.Now().Unix()
	header := s.sign(secret, body, ts)

	s.False(VerifyWebhook(secret, []byte(`{"event_type":"subscription.updated"}`), header))
}

func (s *WebhookSuite) TestVerifyWebhook_expiredTimestamp() {
	secret := "whsec_test"
	body := []byte(`{"event_type":"subscription.created"}`)
	ts := time.Now().Add(-10 * time.Minute).Unix()
	header := s.sign(secret, body, ts)

	s.False(VerifyWebhook(secret, body, header))
}

func (s *WebhookSuite) TestVerifyWebhook_missingInputs() {
	s.False(VerifyWebhook("", []byte("body"), "ts=1;h1=abc"))
	s.False(VerifyWebhook("secret", nil, "ts=1;h1=abc"))
	s.False(VerifyWebhook("secret", []byte("body"), ""))
}

func (s *WebhookSuite) TestParsePaddleSignature() {
	ts, h1, ok := parsePaddleSignature("ts=123;h1=abc")
	s.True(ok)
	s.Equal("123", ts)
	s.Equal("abc", h1)

	_, _, ok = parsePaddleSignature("h1=abc")
	s.False(ok)
}
