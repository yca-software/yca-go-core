package yca_paddle_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	yca_paddle "github.com/yca-software/yca-go-core/paddle"
)

type ModuleSuite struct {
	suite.Suite
}

func TestModuleSuite(t *testing.T) {
	suite.Run(t, new(ModuleSuite))
}

func (s *ModuleSuite) TestNew_requiresAPIKey() {
	_, err := yca_paddle.New(yca_paddle.Config{})
	s.Require().Error(err)
	s.Contains(err.Error(), "API key is required")
}

func (s *ModuleSuite) TestNew_sandbox() {
	mod, err := yca_paddle.New(yca_paddle.Config{
		APIKey:      "test_api_key_123",
		Environment: yca_paddle.PaddleEnvironmentSandbox,
	})
	s.Require().NoError(err)
	s.NotNil(mod.Customer)
	s.NotNil(mod.Subscription)
	s.NotNil(mod.Transaction)
}

func (s *ModuleSuite) TestNew_production() {
	mod, err := yca_paddle.New(yca_paddle.Config{
		APIKey:      "test_api_key_123",
		Environment: yca_paddle.PaddleEnvironmentProduction,
	})
	s.Require().NoError(err)
	s.NotNil(mod.Customer)
	s.NotNil(mod.Subscription)
	s.NotNil(mod.Transaction)
}
