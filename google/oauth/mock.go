package chi_google_oauth

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockOAuth is a testify mock for OAuth.
type MockOAuth struct {
	mock.Mock
}

func (m *MockOAuth) GetUserInfo(ctx context.Context, code string) (*UserInfo, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserInfo), args.Error(1)
}
