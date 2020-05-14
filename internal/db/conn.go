package db

import (
	"time"

	"github.com/ashirko/tcpmirror/internal/monitoring"
	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

// Conn is a type for connection to DB
type Conn redis.Conn

// Connect creates connection to DB
func Connect(dbAddress string, options *util.Options) Conn {
	c := connect(dbAddress)
	monitoring.NewRedisConn(options)
	return c
}

// Close closes connection to DB
func Close(c Conn, options *util.Options) {
	if err := c.Close(); err != nil {
		logrus.Errorf("can't close connection to redis: %s", err)
	}
	monitoring.DelRedisConn(options)
}

func connect(dbAddress string) redis.Conn {
	var cR redis.Conn
	for {
		var err error
		cR, err = redis.Dial("tcp", string(dbAddress))
		if err != nil {
			logrus.Errorf("error connecting to redis: %s\n", err)
		} else {
			break
		}
		time.Sleep(5 * time.Second)
	}
	return cR
}
