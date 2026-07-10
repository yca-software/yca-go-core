package yca_observer_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	yca_observer "github.com/yca-software/yca-go-core/observer"
)

type RateLimitSuite struct {
	suite.Suite
}

func TestRateLimitSuite(t *testing.T) {
	suite.Run(t, new(RateLimitSuite))
}

func (s *RateLimitSuite) TestHashRateLimitIdentifier_stableAndNonReversible() {
	a := yca_observer.HashRateLimitIdentifier("11111111-1111-4111-8111-111111111101")
	b := yca_observer.HashRateLimitIdentifier("11111111-1111-4111-8111-111111111101")
	c := yca_observer.HashRateLimitIdentifier("22222222-2222-4222-8222-222222222202")

	s.Equal(a, b)
	s.NotEqual(a, c)
	s.NotEqual("11111111-1111-4111-8111-111111111101", a)
	s.Len(a, 16)
}

func (s *RateLimitSuite) TestHashRateLimitIdentifier_unknown() {
	s.Equal("unknown", yca_observer.HashRateLimitIdentifier(""))
	s.Equal("unknown", yca_observer.HashRateLimitIdentifier("unknown"))
}
