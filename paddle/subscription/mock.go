package yca_paddle_subscription

import (
	"context"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/stretchr/testify/mock"
)

type MockSubscriptionService struct {
	mock.Mock
}

func (m *MockSubscriptionService) CancelSubscription(ctx context.Context, subscriptionID string) (*paddle.Subscription, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paddle.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) UpdateSubscriptionItems(ctx context.Context, subscriptionID, priceID string) (*paddle.Subscription, error) {
	args := m.Called(ctx, subscriptionID, priceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paddle.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) GetSubscription(ctx context.Context, subscriptionID string) (*paddle.Subscription, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paddle.Subscription), args.Error(1)
}
