package chi_paddle_transaction

import (
	"context"
	"errors"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
)

type TransactionService interface {
	GetTransaction(ctx context.Context, transactionID string) (*paddle.Transaction, error)
	CreateCheckoutSession(ctx context.Context, customerID, priceID string) (*CheckoutSessionResult, error)
}

type transactionService struct {
	sdk *paddle.SDK
}

func New(sdk *paddle.SDK) TransactionService {
	return &transactionService{sdk: sdk}
}

// GetTransaction loads a transaction by ID.
func (s *transactionService) GetTransaction(ctx context.Context, transactionID string) (*paddle.Transaction, error) {
	return s.sdk.TransactionsClient.GetTransaction(ctx, &paddle.GetTransactionRequest{TransactionID: transactionID})
}

// CreateCheckoutSession starts a Paddle transaction for the given customer and price.
func (s *transactionService) CreateCheckoutSession(ctx context.Context, customerID, priceID string) (*CheckoutSessionResult, error) {
	if customerID == "" {
		return nil, errors.New("paddle: customer id required")
	}
	if priceID == "" {
		return nil, errors.New("paddle: price id required")
	}

	transaction, err := s.sdk.TransactionsClient.CreateTransaction(ctx, &paddle.CreateTransactionRequest{
		CustomerID: &customerID,
		Items: []paddle.CreateTransactionItems{
			*paddle.NewCreateTransactionItemsTransactionItemFromCatalog(&paddle.TransactionItemFromCatalog{
				PriceID:  priceID,
				Quantity: 1,
			}),
		},
	})
	if err != nil {
		return nil, err
	}
	if transaction.ID == "" {
		return nil, errors.New("paddle: empty transaction id")
	}
	return &CheckoutSessionResult{TransactionID: transaction.ID}, nil
}
