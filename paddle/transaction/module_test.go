package chi_paddle_transaction

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TransactionModuleSuite struct {
	suite.Suite
}

func TestTransactionModuleSuite(t *testing.T) {
	suite.Run(t, new(TransactionModuleSuite))
}

func (s *TransactionModuleSuite) TestCreateCheckoutSession_validation() {
	svc := &transactionService{}

	_, err := svc.CreateCheckoutSession(context.Background(), "", "pri_123")
	s.Require().Error(err)
	s.Contains(err.Error(), "customer id required")

	_, err = svc.CreateCheckoutSession(context.Background(), "ctm_123", "")
	s.Require().Error(err)
	s.Contains(err.Error(), "price id required")
}
