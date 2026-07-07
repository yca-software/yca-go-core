package chi_paddle_customer

import (
	"errors"
	"regexp"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/PaddleHQ/paddle-go-sdk/v4/pkg/paddleerr"
)

var paddleCustomerIDFromMessageRE = regexp.MustCompile(`\b(ctm_[0-9a-z]+)\b`)

func isCustomerAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, paddle.ErrCustomerAlreadyExists) {
		return true
	}
	var pe *paddleerr.Error
	if errors.As(err, &pe) && pe != nil && pe.Code == "customer_already_exists" {
		return true
	}
	return false
}

func paddleCustomerIDFromAlreadyExistsError(err error) string {
	if err == nil {
		return ""
	}
	var pe *paddleerr.Error
	if errors.As(err, &pe) && pe != nil {
		if id := extractPaddleCustomerIDFromText(pe.Detail); id != "" {
			return id
		}
	}
	return extractPaddleCustomerIDFromText(err.Error())
}

func extractPaddleCustomerIDFromText(s string) string {
	if s == "" {
		return ""
	}
	m := paddleCustomerIDFromMessageRE.FindStringSubmatch(s)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}
