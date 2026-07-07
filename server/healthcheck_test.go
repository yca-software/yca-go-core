package chi_server

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type stubReadinessDep struct {
	err   error
	delay time.Duration
	calls atomic.Int32
}

func (d *stubReadinessDep) Check(ctx context.Context) error {
	d.calls.Add(1)
	if d.delay == 0 {
		return d.err
	}
	select {
	case <-time.After(d.delay):
		return d.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

type ReadinessSuite struct {
	suite.Suite
}

func TestReadinessSuite(t *testing.T) {
	suite.Run(t, new(ReadinessSuite))
}

func (s *ReadinessSuite) TestNoDependencies() {
	err := checkDependenciesConcurrently(context.Background(), nil)
	s.NoError(err)
}

func (s *ReadinessSuite) TestAllPass() {
	err := checkDependenciesConcurrently(context.Background(), []ReadinessDependency{
		&stubReadinessDep{},
		&stubReadinessDep{},
	})
	s.NoError(err)
}

func (s *ReadinessSuite) TestFailFast() {
	slow := &stubReadinessDep{delay: 200 * time.Millisecond}
	fastFail := &stubReadinessDep{err: errors.New("cache unavailable")}

	start := time.Now()
	err := checkDependenciesConcurrently(context.Background(), []ReadinessDependency{slow, fastFail})
	elapsed := time.Since(start)

	s.EqualError(err, "cache unavailable")
	s.Less(elapsed, 150*time.Millisecond, "should fail fast without waiting for slow check")
}

func (s *ReadinessSuite) TestCancelsRemainingChecks() {
	slow := &stubReadinessDep{delay: 500 * time.Millisecond}
	fastFail := &stubReadinessDep{err: errors.New("down")}

	_ = checkDependenciesConcurrently(context.Background(), []ReadinessDependency{slow, fastFail})

	s.Equal(int32(1), slow.calls.Load())
}
