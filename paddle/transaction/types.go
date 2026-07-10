package yca_paddle_transaction

// CheckoutSessionResult is returned after creating a checkout transaction.
type CheckoutSessionResult struct {
	TransactionID string `json:"transactionId"`
}
