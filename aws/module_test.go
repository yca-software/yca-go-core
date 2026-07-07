package chi_aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	chi_aws_iot "github.com/yca-software/yca-go-core/aws/iot"
	chi_aws_ses "github.com/yca-software/yca-go-core/aws/ses"
)

type ModuleSuite struct {
	suite.Suite
}

func TestModuleSuite(t *testing.T) {
	suite.Run(t, new(ModuleSuite))
}

func (s *ModuleSuite) TestNew_DisabledServices() {
	mod, err := New(context.Background(), Config{})
	s.Require().NoError(err)
	s.Nil(mod.SES)
	s.Nil(mod.SQS)
	s.Nil(mod.S3)
	s.Nil(mod.IoT)
}

func (s *ModuleSuite) TestLoadSDKConfig_UsesRootConfigByDefault() {
	awsCfg, err := loadSDKConfig(context.Background(), Config{
		Region:   "eu-central-1",
		Endpoint: "http://localhost:4566",
	}, serviceSQS)
	s.Require().NoError(err)
	s.Equal("eu-central-1", awsCfg.Region)
	s.Require().NotNil(awsCfg.BaseEndpoint)
	s.Equal("http://localhost:4566", *awsCfg.BaseEndpoint)
}

func (s *ModuleSuite) TestLoadSDKConfig_SESOverridesRootConfig() {
	awsCfg, err := loadSDKConfig(context.Background(), Config{
		Region:   "eu-central-1",
		Endpoint: "http://localhost:4566",
		SES: &chi_aws_ses.Config{
			Region:   "us-east-1",
			Endpoint: "http://localhost:4570",
		},
	}, serviceSES)
	s.Require().NoError(err)
	s.Equal("us-east-1", awsCfg.Region)
	s.Require().NotNil(awsCfg.BaseEndpoint)
	s.Equal("http://localhost:4570", *awsCfg.BaseEndpoint)
}

func (s *ModuleSuite) TestLoadSDKConfig_IoTUsesIoTRegion() {
	awsCfg, err := loadSDKConfig(context.Background(), Config{
		Region: "us-east-1",
		IoT: &chi_aws_iot.Config{
			Region: "eu-west-1",
		},
	}, serviceIoT)
	s.Require().NoError(err)
	s.Equal("eu-west-1", awsCfg.Region)
}
