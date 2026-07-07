package chi_aws_ses

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SESModuleSuite struct {
	suite.Suite
}

func TestSESModuleSuite(t *testing.T) {
	suite.Run(t, new(SESModuleSuite))
}

func (s *SESModuleSuite) TestConfigEnabled() {
	s.False((Config{}).Enabled())
	s.True((Config{Region: "eu-central-1"}).Enabled())
	s.True((Config{Endpoint: "http://localhost:4566"}).Enabled())
	s.True((Config{FromEmail: "noreply@example.com"}).Enabled())
}

func (s *SESModuleSuite) TestFormatFromAddress() {
	s.Equal("noreply@example.com", formatFromAddress("noreply@example.com", ""))
	s.Equal("2Chi <noreply@example.com>", formatFromAddress("noreply@example.com", "2Chi"))
}
