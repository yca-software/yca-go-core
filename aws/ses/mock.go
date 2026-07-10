package yca_aws_ses

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockSES struct {
	mock.Mock
}

func (m *MockSES) Send(ctx context.Context, msg SESEmailDataPayload) error {
	return m.Called(ctx, msg).Error(0)
}
