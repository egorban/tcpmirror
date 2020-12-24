package db

import (
	"time"

	"github.com/ashirko/tcpmirror/internal/monitoring"
	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

// Pool defines pool of connections to DB
type Pool struct {
	*redis.Pool
}

// NewPool create new pool of connections to DB
func NewPool(dbAddress string, options *util.Options) (pool *Pool) {
	return newPool(dbAddress, options)
}

// Close closes pull of connections to DB
func (pool Pool) Close() error {
	return pool.Pool.Close()
}

func newPool(addr string, options *util.Options) *Pool {
	r := &redis.Pool{
		MaxIdle:     20,
		MaxActive:   0,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			logrus.Infoln("redisConn pool")
			monitoring.SendMetricInfo(options, monitoring.RedisConnPool, monitoring.TypeRedis)
			return redis.Dial("tcp", addr)
		},
	}
	return &Pool{
		Pool: r,
	}
}
