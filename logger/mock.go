package yca_logger

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockLogger is a testify-based test double for [Logger], in the same style
// as go-common: embed [mock.Mock] and register expectations with [mock.Mock.On].
//
// Variadic log methods record msg as the first argument, then each attr pair.
// [MockLogger.With] and [MockLogger.WithContext] must use .Return(Logger)
// on expectations when those methods are exercised, for example:
//
//	m := new(MockLogger)
//	m.On("Info", "hello", "k", 1).Return()
//	m.On("With", "svc", "api").Return(m).Maybe()
//	m.On("WithContext", mock.Anything).Return(m).Maybe()
//
// Prefer [NewStubMockLogger] when the code under test may log but the test does
// not assert on log output — it accepts any call without per-method stubs.
type MockLogger struct {
	mock.Mock
	stub bool
}

// NewStubMockLogger returns a [MockLogger] that accepts any With / WithContext /
// Debug / Info / Warn / Error call. Use this in service/handler suites instead
// of copying Maybe() stubs into each package.
func NewStubMockLogger() *MockLogger {
	return &MockLogger{stub: true}
}

func (m *MockLogger) Debug(msg string, args ...any) {
	if m.stub {
		return
	}
	m.Called(append([]any{msg}, args...)...)
}

func (m *MockLogger) Info(msg string, args ...any) {
	if m.stub {
		return
	}
	m.Called(append([]any{msg}, args...)...)
}

func (m *MockLogger) Warn(msg string, args ...any) {
	if m.stub {
		return
	}
	m.Called(append([]any{msg}, args...)...)
}

func (m *MockLogger) Error(msg string, args ...any) {
	if m.stub {
		return
	}
	m.Called(append([]any{msg}, args...)...)
}

func (m *MockLogger) With(args ...any) Logger {
	if m.stub {
		return m
	}
	ret := m.Called(args...)
	return ret.Get(0).(Logger)
}

func (m *MockLogger) WithContext(ctx context.Context) Logger {
	if m.stub {
		return m
	}
	ret := m.Called(ctx)
	return ret.Get(0).(Logger)
}
