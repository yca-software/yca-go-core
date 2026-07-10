package yca_aws

import (
	"context"

	aws_sdk "github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	yca_aws_iot "github.com/yca-software/yca-go-core/aws/iot"
	yca_aws_s3 "github.com/yca-software/yca-go-core/aws/s3"
	yca_aws_ses "github.com/yca-software/yca-go-core/aws/ses"
	yca_aws_sqs "github.com/yca-software/yca-go-core/aws/sqs"
)

const (
	serviceSES = "ses"
	serviceSQS = "sqs"
	serviceS3  = "s3"
	serviceIoT = "iot"
)

type Config struct {
	Region   string
	Endpoint string

	SES *yca_aws_ses.Config
	SQS *yca_aws_sqs.Config
	S3  *yca_aws_s3.Config
	IoT *yca_aws_iot.Config
}

type Module struct {
	SES yca_aws_ses.SES
	SQS yca_aws_sqs.SQS
	S3  yca_aws_s3.S3
	IoT yca_aws_iot.IoT
}

func New(ctx context.Context, cfg Config) (*Module, error) {
	m := &Module{}

	if cfg.SES != nil && cfg.SES.Enabled() {
		awsCfg, err := loadSDKConfig(ctx, cfg, serviceSES)
		if err != nil {
			return nil, err
		}
		m.SES = yca_aws_ses.NewSESClient(awsCfg, *cfg.SES)
	}

	if cfg.S3 != nil && cfg.S3.Enabled() {
		awsCfg, err := loadSDKConfig(ctx, cfg, serviceS3)
		if err != nil {
			return nil, err
		}
		m.S3 = yca_aws_s3.NewS3Client(awsCfg)
	}

	if cfg.SQS != nil && cfg.SQS.Enabled() {
		awsCfg, err := loadSDKConfig(ctx, cfg, serviceSQS)
		if err != nil {
			return nil, err
		}
		m.SQS = yca_aws_sqs.NewSQSClient(awsCfg)
	}

	if cfg.IoT != nil && cfg.IoT.Enabled() {
		awsCfg, err := loadSDKConfig(ctx, cfg, serviceIoT)
		if err != nil {
			return nil, err
		}
		iotClient, err := yca_aws_iot.NewIoTClient(awsCfg)
		if err != nil {
			return nil, err
		}
		m.IoT = iotClient
	}

	return m, nil
}

func loadSDKConfig(ctx context.Context, cfg Config, service string) (aws_sdk.Config, error) {
	region := cfg.Region
	endpoint := cfg.Endpoint

	switch service {
	case serviceSES:
		if cfg.SES != nil {
			if cfg.SES.Region != "" {
				region = cfg.SES.Region
			}
			if cfg.SES.Endpoint != "" {
				endpoint = cfg.SES.Endpoint
			}
		}
	case serviceIoT:
		if cfg.IoT != nil && cfg.IoT.Region != "" {
			region = cfg.IoT.Region
		}
	}

	opts := []func(*aws_config.LoadOptions) error{
		aws_config.WithRegion(region),
		aws_config.WithBaseEndpoint(endpoint),
	}

	return aws_config.LoadDefaultConfig(ctx, opts...)
}
