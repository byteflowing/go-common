package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/byteflowing/go-common/idx"
)

const (
	defaultKeyPrefix    = "lock"
	defaultTries        = 5
	defaultWaitDuration = 5 * time.Millisecond
)

func (r *Redis) Lock(ctx context.Context, key string, expiration time.Duration, options ...Option) (identifier string, err error) {
	identifier = idx.UUIDv4()
	prefix, waitDuration, tries := parseLockOptions(options)
	key = fmt.Sprintf("%s:%s", prefix, key)
	for i := 0; i < tries; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			success, err := r.SetNX(ctx, key, identifier, expiration).Result()
			if err != nil {
				return "", err
			}
			if success {
				return identifier, nil
			}
			time.Sleep(waitDuration)
		}
	}
	return "", errors.New("failed to acquire lock within tries")
}

func (r *Redis) Unlock(ctx context.Context, key, identifier string, options ...Option) (err error) {
	prefix, _, _ := parseLockOptions(options)
	key = fmt.Sprintf("%s:%s", prefix, key)
	var command = redis.NewScript(`
		if
			redis.call("GET", KEYS[1]) == ARGV[1]
		then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)
	result, err := command.Run(ctx, r, []string{key}, []string{identifier}).Int()
	if err != nil {
		return err
	}
	if result == 1 {
		return nil
	}
	return fmt.Errorf("unlock failed: key %s is no longer valid or mismatched identifier", key)
}

func (r *Redis) RenewLock(ctx context.Context, key, identifier string, expiration time.Duration, options ...Option) (err error) {
	prefix, _, _ := parseLockOptions(options)
	key = fmt.Sprintf("%s:%s", prefix, key)
	command := redis.NewScript(`
		if
			redis.call("GET", KEYS[1]) == ARGV[1]
		then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`)
	result, err := command.Run(ctx, r, []string{key}, identifier, int(expiration.Milliseconds())).Int()
	if err != nil {
		return err
	}
	if result == 1 {
		return nil
	}
	return fmt.Errorf("renew failed: key %s is no longer valid or mismatched identifier", key)
}

// WithLockKeyPrefix : 设置key前缀
func WithLockKeyPrefix(prefix string) Option {
	return func(o *Options) {
		o.LockKeyPrefix = prefix
	}
}

// WithLockTryTimes : 设置抢锁失败重试次数
func WithLockTryTimes(times int) Option {
	return func(o *Options) {
		o.LockTryTimes = times
	}
}

// WithLockWaitDuration : 设置两次抢锁直接等待间隔
func WithLockWaitDuration(duration time.Duration) Option {
	return func(o *Options) {
		o.LockWaitDuration = duration
	}
}

func parseLockOptions(options []Option) (prefix string, waitDuration time.Duration, tries int) {
	prefix = defaultKeyPrefix
	waitDuration = defaultWaitDuration
	tries = defaultTries
	if len(options) == 0 {
		return
	}
	ops := &Options{}
	for _, op := range options {
		op(ops)
	}
	if ops.LockTryTimes > 0 {
		tries = ops.LockTryTimes
	}
	if ops.LockWaitDuration > 0 {
		waitDuration = ops.LockWaitDuration
	}
	if len(ops.LockKeyPrefix) > 0 {
		prefix = ops.LockKeyPrefix
	}
	return
}
