package chi_aws_sqs

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockSQS struct {
	mock.Mock
}

func (m *MockSQS) SendMessage(ctx context.Context, queueURL string, body []byte, attributes map[string]string) error {
	return m.Called(ctx, queueURL, body, attributes).Error(0)
}

func (m *MockSQS) ReceiveMessages(ctx context.Context, opts ReceiveOptions) ([]QueueMessage, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]QueueMessage), args.Error(1)
}

func (m *MockSQS) DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error {
	return m.Called(ctx, queueURL, receiptHandle).Error(0)
}

func (m *MockSQS) ChangeMessageVisibility(ctx context.Context, queueURL, receiptHandle string, visibilityTimeout int32) error {
	return m.Called(ctx, queueURL, receiptHandle, visibilityTimeout).Error(0)
}
