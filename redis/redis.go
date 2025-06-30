package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	LockKeyPrefix    string
	LockTryTimes     int
	LockWaitDuration time.Duration
}

type Option func(o *Options)

type Redis struct {
	redis.Cmdable
}

func New(c *Config) *Redis {
	opts := &redis.Options{
		Addr:       c.Addr,
		ClientName: c.ClientName,
		Protocol:   c.Protocol,
		DB:         c.DB,
	}
	if len(c.User) > 0 {
		opts.Username = c.User
	}
	if len(c.Password) > 0 {
		opts.Password = c.Password
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic("connecting to redis failed:" + err.Error())
	}
	return &Redis{rdb}
}
