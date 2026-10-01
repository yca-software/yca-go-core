package yca_logger_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	yca_logger "github.com/yca-software/yca-go-core/logger"
)

func TestNewStubMockLogger_AcceptsAnyCall(t *testing.T) {
	m := yca_logger.NewStubMockLogger()

	require.Same(t, m, m.With("k", "v"))
	require.Same(t, m, m.WithContext(context.Background()))
	require.NotPanics(t, func() {
		m.Debug("d", "a", 1)
		m.Info("i")
		m.Warn("w", "x", true)
		m.Error("e", "err", "boom")
	})
}
