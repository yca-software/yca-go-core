package yca_paddle_customer

import (
	"context"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/stretchr/testify/mock"
)

type MockCustomerService struct {
	mock.Mock
}

func (m *MockCustomerService) CreateCustomer(ctx context.Context, req CustomerInput) (*paddle.Customer, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paddle.Customer), args.Error(1)
}

func (m *MockCustomerService) UpdateCustomer(ctx context.Context, customerID string, req CustomerInput) (*paddle.Customer, error) {
	args := m.Called(ctx, customerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paddle.Customer), args.Error(1)
}

func (m *MockCustomerService) ArchiveCustomer(ctx context.Context, customerID string) error {
	args := m.Called(ctx, customerID)
	return args.Error(0)
}

func (m *MockCustomerService) CreateCustomerPortalSession(ctx context.Context, customerID string) (*CustomerPortalSessionResult, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CustomerPortalSessionResult), args.Error(1)
}
