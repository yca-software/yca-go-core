package yca_token_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	yca_token "github.com/yca-software/yca-go-core/token"
)

type GenerateTestSuite struct {
	suite.Suite
}

func TestGenerateTestSuite(t *testing.T) {
	suite.Run(t, new(GenerateTestSuite))
}

func (s *GenerateTestSuite) TestGenerateOpaqueToken() {
	token, err := yca_token.GenerateOpaqueToken()
	s.NoError(err)
	s.NotEmpty(token)
	s.NotContains(token, "=")
}

func (s *GenerateTestSuite) TestGenerateOpaqueToken_ProducesDifferentTokens() {
	token1, err1 := yca_token.GenerateOpaqueToken()
	s.NoError(err1)

	token2, err2 := yca_token.GenerateOpaqueToken()
	s.NoError(err2)

	s.NotEqual(token1, token2, "Each token should be unique")
}

func (s *GenerateTestSuite) TestGenerateOpaqueToken_NoPadding() {
	token, err := yca_token.GenerateOpaqueToken()
	s.NoError(err)
	s.NotEmpty(token)
	s.NotContains(token, "=")
}
