package yca_aws_iot

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type IoTModuleSuite struct {
	suite.Suite
}

func TestIoTModuleSuite(t *testing.T) {
	suite.Run(t, new(IoTModuleSuite))
}

func (s *IoTModuleSuite) TestConfigEnabled() {
	s.False((Config{}).Enabled())
	s.True((Config{Region: "eu-central-1"}).Enabled())
}
