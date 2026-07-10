package yca_server_test

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/suite"

	yca_server "github.com/yca-software/yca-go-core/server"
)

type CleanupSuite struct {
	suite.Suite
}

func TestCleanupSuite(t *testing.T) {
	suite.Run(t, new(CleanupSuite))
}

type stubCleanup struct {
	cleanups atomic.Int32
}

func (s *stubCleanup) Cleanup() { s.cleanups.Add(1) }

func (s *CleanupSuite) TestCleanupDependenciesConcurrently() {
	a := &stubCleanup{}
	b := &stubCleanup{}
	yca_server.CleanupDependenciesConcurrently([]yca_server.CleanupDependency{a, b})
	s.Equal(int32(1), a.cleanups.Load())
	s.Equal(int32(1), b.cleanups.Load())
}
