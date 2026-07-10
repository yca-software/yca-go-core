package yca_paddle_customer

import (
	"errors"
	"testing"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/PaddleHQ/paddle-go-sdk/v4/pkg/paddleerr"
	"github.com/stretchr/testify/suite"
)

type CustomerHelpersSuite struct {
	suite.Suite
}

func TestCustomerHelpersSuite(t *testing.T) {
	suite.Run(t, new(CustomerHelpersSuite))
}

func (s *CustomerHelpersSuite) TestIsCustomerAlreadyExistsError() {
	s.False(isCustomerAlreadyExistsError(nil))
	s.False(isCustomerAlreadyExistsError(errors.New("other")))

	s.True(isCustomerAlreadyExistsError(paddle.ErrCustomerAlreadyExists))
	s.True(isCustomerAlreadyExistsError(&paddleerr.Error{Code: "customer_already_exists"}))
}

func (s *CustomerHelpersSuite) TestPaddleCustomerIDFromAlreadyExistsError() {
	s.Equal("", paddleCustomerIDFromAlreadyExistsError(nil))

	err := &paddleerr.Error{
		Code:   "customer_already_exists",
		Detail: "Customer ctm_01h2x3y4z5 already exists for email.",
	}
	s.Equal("ctm_01h2x3y4z5", paddleCustomerIDFromAlreadyExistsError(err))
}

func (s *CustomerHelpersSuite) TestExtractPaddleCustomerIDFromText() {
	s.Equal("", extractPaddleCustomerIDFromText(""))
	s.Equal("ctm_abc123", extractPaddleCustomerIDFromText("duplicate customer ctm_abc123"))
}
