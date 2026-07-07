package chi_paddle_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	chi_paddle "github.com/yca-software/yca-go-core/paddle"
)

type ModuleSuite struct {
	suite.Suite
}

func TestModuleSuite(t *testing.T) {
	suite.Run(t, new(ModuleSuite))
}

func (s *ModuleSuite) TestNew_requiresAPIKey() {
	_, err := chi_paddle.New(chi_paddle.Config{})
	s.Require().Error(err)
	s.Contains(err.Error(), "API key is required")
}

func (s *ModuleSuite) TestNew_sandbox() {
	mod, err := chi_paddle.New(chi_paddle.Config{
		APIKey:      "test_api_key_123",
		Environment: chi_paddle.PaddleEnvironmentSandbox,
	})
	s.Require().NoError(err)
	s.NotNil(mod.Customer)
	s.NotNil(mod.Subscription)
	s.NotNil(mod.Transaction)
}

func (s *ModuleSuite) TestNew_production() {
	mod, err := chi_paddle.New(chi_paddle.Config{
		APIKey:      "test_api_key_123",
		Environment: chi_paddle.PaddleEnvironmentProduction,
	})
	s.Require().NoError(err)
	s.NotNil(mod.Customer)
	s.NotNil(mod.Subscription)
	s.NotNil(mod.Transaction)
}
