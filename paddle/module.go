package yca_paddle

import (
	"errors"
	"fmt"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	yca_paddle_customer "github.com/yca-software/yca-go-core/paddle/customer"
	yca_paddle_subscription "github.com/yca-software/yca-go-core/paddle/subscription"
	yca_paddle_transaction "github.com/yca-software/yca-go-core/paddle/transaction"
)

type PaddleEnvironment string

const (
	PaddleEnvironmentSandbox    PaddleEnvironment = "sandbox"
	PaddleEnvironmentProduction PaddleEnvironment = "production"
)

type Config struct {
	APIKey      string
	Environment PaddleEnvironment
}

type Module struct {
	Customer     yca_paddle_customer.CustomerService
	Subscription yca_paddle_subscription.SubscriptionService
	Transaction  yca_paddle_transaction.TransactionService
}

func New(cfg Config) (*Module, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("paddle: API key is required")
	}

	baseURL := paddle.SandboxBaseURL
	if cfg.Environment == PaddleEnvironmentProduction {
		baseURL = paddle.ProductionBaseURL
	}

	sdk, err := paddle.New(cfg.APIKey, paddle.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("paddle: initialize SDK: %w", err)
	}

	return &Module{
		Customer:     yca_paddle_customer.New(sdk),
		Subscription: yca_paddle_subscription.New(sdk),
		Transaction:  yca_paddle_transaction.New(sdk),
	}, nil
}
