package chi_paddle_transaction

import (
	"context"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/stretchr/testify/mock"
)

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) GetTransaction(ctx context.Context, transactionID string) (*paddle.Transaction, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paddle.Transaction), args.Error(1)
}

func (m *MockTransactionService) CreateCheckoutSession(ctx context.Context, customerID, priceID string) (*CheckoutSessionResult, error) {
	args := m.Called(ctx, customerID, priceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CheckoutSessionResult), args.Error(1)
}
