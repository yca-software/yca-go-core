package yca_aws_iot

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConnectivityTimestampIsReliable(t *testing.T) {
	t.Parallel()

	require.False(t, ConnectivityTimestampIsReliable(time.Time{}))
	require.False(t, ConnectivityTimestampIsReliable(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)))
	require.True(t, ConnectivityTimestampIsReliable(time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)))
}
