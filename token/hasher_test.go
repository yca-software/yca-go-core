package chi_token_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	chi_token "github.com/yca-software/yca-go-core/token"
)

type HasherTestSuite struct {
	suite.Suite
}

func TestHasherTestSuite(t *testing.T) {
	suite.Run(t, new(HasherTestSuite))
}

func (s *HasherTestSuite) TestNewHasher_RequiresPepper() {
	s.Panics(func() {
		chi_token.NewHasher("")
	})
}

func (s *HasherTestSuite) TestHasher_Hash_Consistent() {
	hasher := chi_token.NewHasher("pepper-secret")
	plain := "refresh-token"
	s.Equal(hasher.Hash(plain), hasher.Hash(plain))
	s.Len(hasher.Hash(plain), 64)
}

func (s *HasherTestSuite) TestHasher_Hash_DifferentTokens() {
	hasher := chi_token.NewHasher("pepper-secret")
	s.NotEqual(hasher.Hash("token-a"), hasher.Hash("token-b"))
}

func (s *HasherTestSuite) TestHasher_Hash_DifferentPeppers() {
	plain := "opaque-token-value"
	h1 := chi_token.NewHasher("pepper-a")
	h2 := chi_token.NewHasher("pepper-b")
	s.NotEqual(h1.Hash(plain), h2.Hash(plain))
}

func (s *HasherTestSuite) TestHasher_Verify() {
	hasher := chi_token.NewHasher("pepper-secret")
	plain := "reset-token"
	s.True(hasher.Verify(hasher.Hash(plain), plain))
	s.False(hasher.Verify("deadbeef", plain))
}
