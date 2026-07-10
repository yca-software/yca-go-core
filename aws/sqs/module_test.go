package yca_aws_sqs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SQSModuleSuite struct {
	suite.Suite
}

func TestSQSModuleSuite(t *testing.T) {
	suite.Run(t, new(SQSModuleSuite))
}

func (s *SQSModuleSuite) TestConfigEnabled() {
	s.False((Config{}).Enabled())
	s.True((Config{Region: "eu-central-1"}).Enabled())
	s.True((Config{Endpoint: "http://localhost:4566"}).Enabled())
}
