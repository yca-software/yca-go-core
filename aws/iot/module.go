package chi_aws_iot

import (
	"context"
	"fmt"
	"strings"

	aws_sdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/aws/aws-sdk-go-v2/service/iotdataplane"
)

const iotDataEndpointType = "iot:Data-ATS"

type Config struct {
	Region string
}

func (c Config) Enabled() bool {
	return c.Region != ""
}

type IoT interface {
	GetThingShadow(ctx context.Context, thingName string) ([]byte, error)
	UpdateThingShadow(ctx context.Context, thingName string, payload []byte) ([]byte, error)
}

type iotClient struct {
	dataplane *iotdataplane.Client
}

func NewIoTClient(cfg aws_sdk.Config) (IoT, error) {
	ctx := context.Background()
	awsIoTClient := iot.NewFromConfig(cfg)

	desc, err := awsIoTClient.DescribeEndpoint(ctx, &iot.DescribeEndpointInput{
		EndpointType: aws_sdk.String(iotDataEndpointType),
	})
	if err != nil {
		return nil, fmt.Errorf("aws/iot: describe endpoint: %w", err)
	}
	if desc.EndpointAddress == nil || strings.TrimSpace(*desc.EndpointAddress) == "" {
		return nil, fmt.Errorf("aws/iot: empty data plane endpoint address")
	}
	dataEndpoint := strings.TrimSpace(*desc.EndpointAddress)
	if !strings.HasPrefix(dataEndpoint, "https://") {
		dataEndpoint = "https://" + dataEndpoint
	}

	dataplaneCfg := cfg
	dataplaneCfg.BaseEndpoint = aws_sdk.String(dataEndpoint)
	dataplane := iotdataplane.NewFromConfig(dataplaneCfg)

	return &iotClient{dataplane: dataplane}, nil
}

func (c *iotClient) GetThingShadow(ctx context.Context, thingName string) ([]byte, error) {
	if thingName == "" {
		return nil, fmt.Errorf("aws/iot: thing name is required")
	}
	out, err := c.dataplane.GetThingShadow(ctx, &iotdataplane.GetThingShadowInput{
		ThingName: aws_sdk.String(thingName),
	})
	if err != nil {
		return nil, fmt.Errorf("aws/iot: get thing shadow for %s: %w", thingName, err)
	}
	return out.Payload, nil
}

func (c *iotClient) UpdateThingShadow(ctx context.Context, thingName string, payload []byte) ([]byte, error) {
	if thingName == "" {
		return nil, fmt.Errorf("aws/iot: thing name is required")
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("aws/iot: shadow payload is required")
	}
	out, err := c.dataplane.UpdateThingShadow(ctx, &iotdataplane.UpdateThingShadowInput{
		ThingName: aws_sdk.String(thingName),
		Payload:   payload,
	})
	if err != nil {
		return nil, fmt.Errorf("aws/iot: update thing shadow for %s: %w", thingName, err)
	}
	return out.Payload, nil
}
