package yca_aws_iot

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockIoT struct {
	mock.Mock
}

func (m *MockIoT) GetThingShadow(ctx context.Context, thingName string) ([]byte, error) {
	args := m.Called(ctx, thingName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockIoT) UpdateThingShadow(ctx context.Context, thingName string, payload []byte) ([]byte, error) {
	args := m.Called(ctx, thingName, payload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockIoT) GetThingConnectivityData(ctx context.Context, thingName string) (*ThingConnectivityData, error) {
	args := m.Called(ctx, thingName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ThingConnectivityData), args.Error(1)
}
