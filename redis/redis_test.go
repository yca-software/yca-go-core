package yca_redis_test

import (
	"context"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/suite"

	yca_redis "github.com/yca-software/yca-go-core/redis"
)

type RedisSuite struct {
	suite.Suite
	mr       *miniredis.Miniredis
	redisDSN string
}

func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(RedisSuite))
}

func (s *RedisSuite) SetupSuite() {
	mr, err := miniredis.Run()
	s.Require().NoError(err)
	s.mr = mr
	s.redisDSN = "redis://" + mr.Addr() + "/0"
}

func (s *RedisSuite) TearDownSuite() {
	if s.mr == nil {
		return
	}
	s.mr.Close()
	s.mr = nil
}

func (s *RedisSuite) TestNewRedis_success() {
	r, err := yca_redis.NewRedis(yca_redis.RedisClientConfig{DSN: s.redisDSN})
	s.Require().NoError(err)
	s.Require().NotNil(r)
	s.Require().NotNil(r.GetClient())

	ctx := context.Background()
	s.NoError(r.Check(ctx))

	r.Cleanup()
	s.Error(r.Check(ctx))
}

func (s *RedisSuite) TestNewRedis_invalidDSN() {
	r, err := yca_redis.NewRedis(yca_redis.RedisClientConfig{DSN: "::not-a-redis-url"})
	s.Error(err)
	s.Nil(r)
}
