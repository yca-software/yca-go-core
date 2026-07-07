package chi_paddle

import (
	"errors"
	"fmt"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	chi_paddle_customer "github.com/yca-software/yca-go-core/paddle/customer"
	chi_paddle_subscription "github.com/yca-software/yca-go-core/paddle/subscription"
	chi_paddle_transaction "github.com/yca-software/yca-go-core/paddle/transaction"
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
	Customer     chi_paddle_customer.CustomerService
	Subscription chi_paddle_subscription.SubscriptionService
	Transaction  chi_paddle_transaction.TransactionService
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
		Customer:     chi_paddle_customer.New(sdk),
		Subscription: chi_paddle_subscription.New(sdk),
		Transaction:  chi_paddle_transaction.New(sdk),
	}, nil
}
