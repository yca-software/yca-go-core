# 2Chi Go Paddle

Paddle Billing integration for 2Chi services.

```go
import chi_paddle "github.com/yca-software/yca-go-core/paddle"
```

Subpackages expose service interfaces, types, and mocks:

```go
import (
    chi_paddle_customer "github.com/yca-software/yca-go-core/paddle/customer"
    chi_paddle_subscription "github.com/yca-software/yca-go-core/paddle/subscription"
    chi_paddle_transaction "github.com/yca-software/yca-go-core/paddle/transaction"
)
```

## Setup

```go
mod, err := chi_paddle.New(chi_paddle.Config{
    APIKey:      cfg.PaddleAPIKey,
    Environment: chi_paddle.PaddleEnvironmentSandbox, // or PaddleEnvironmentProduction
})
if err != nil {
    return err
}
```

## Services

| Subpackage | Interface | Capabilities |
| --- | --- | --- |
| `customer` | `CustomerService` | Create/update/archive customers, customer portal sessions |
| `subscription` | `SubscriptionService` | Get, cancel, update subscription items |
| `transaction` | `TransactionService` | Get transaction, create checkout session |

Each subpackage includes a testify mock (`MockCustomerService`, `MockSubscriptionService`, `MockTransactionService`).

## Webhooks

```go
if !chi_paddle.VerifyWebhook(webhookSecret, body, r.Header.Get("Paddle-Signature")) {
    return http.StatusUnauthorized
}

var event chi_paddle.WebhookEvent
// decode body into event
```

`VerifyWebhook` checks the `Paddle-Signature` header (`ts=...;h1=...`) against the raw request body.

## Tests

```bash
go test -race -count=1 ./...
```

Tests are unit-level and do not call the live Paddle API.
