package yca_aws_s3

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

type MockS3 struct {
	mock.Mock
}

func (m *MockS3) PutObject(ctx context.Context, opts PutObjectOptions) (string, error) {
	args := m.Called(ctx, opts)
	return args.String(0), args.Error(1)
}

func (m *MockS3) GetObject(ctx context.Context, opts GetObjectOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockS3) PresignGetObject(ctx context.Context, opts PresignGetObjectOptions) (string, error) {
	args := m.Called(ctx, opts)
	return args.String(0), args.Error(1)
}

func (m *MockS3) DeleteObject(ctx context.Context, bucket, key string) error {
	return m.Called(ctx, bucket, key).Error(0)
}
