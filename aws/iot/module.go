package yca_aws_iot

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	GetThingConnectivityData(ctx context.Context, thingName string) (*ThingConnectivityData, error)
}

type iotClient struct {
	control   *iot.Client
	dataplane *iotdataplane.Client
}

func NewIoTClient(cfg aws_sdk.Config) (IoT, error) {
	ctx := context.Background()
	control := iot.NewFromConfig(cfg)

	desc, err := control.DescribeEndpoint(ctx, &iot.DescribeEndpointInput{
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

	return &iotClient{control: control, dataplane: dataplane}, nil
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

// GetThingConnectivityData queries AWS IoT Core fleet indexing for thing MQTT connectivity.
// Fleet indexing must be enabled on the account before this API succeeds.
func (c *iotClient) GetThingConnectivityData(ctx context.Context, thingName string) (*ThingConnectivityData, error) {
	if thingName == "" {
		return nil, fmt.Errorf("aws/iot: thing name is required")
	}
	out, err := c.control.GetThingConnectivityData(ctx, &iot.GetThingConnectivityDataInput{
		ThingName: aws_sdk.String(thingName),
	})
	if err != nil {
		return nil, fmt.Errorf("aws/iot: get thing connectivity for %s: %w", thingName, err)
	}

	data := &ThingConnectivityData{}
	if out.Connected != nil {
		data.Connected = *out.Connected
	}
	if out.Timestamp != nil {
		data.Timestamp = *out.Timestamp
	}
	if out.DisconnectReason != "" {
		data.DisconnectReason = string(out.DisconnectReason)
	}
	if out.ClientId != nil {
		data.ClientID = *out.ClientId
	}
	return data, nil
}

// ConnectivityTimestampIsReliable reports whether IoT returned a usable event timestamp.
func ConnectivityTimestampIsReliable(ts time.Time) bool {
	return !ts.IsZero() && ts.Year() >= 2000
}
