package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/byteflowing/go-common/idx"
	"github.com/redis/go-redis/v9"
)

const (
	defaultKeyPrefix    = "lock"
	defaultTries        = 5
	defaultWaitDuration = 5 * time.Millisecond
)

type Options struct {
	LockKeyPrefix    string
	LockTryTimes     int
	LockWaitDuration time.Duration
}

type Option func(o *Options)

type Redis struct {
	redis.Cmdable
	_unLockOnce             sync.Once
	_renewLockOnce          sync.Once
	_allowFixedLimitOnce    sync.Once
	_incrWithExpireLockOnce sync.Once
	_unLockScript           *redis.Script
	_renewLockScript        *redis.Script
	_allowFixedLimitScript  *redis.Script
	_incrWithExpireScript   *redis.Script
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
	return &Redis{Cmdable: rdb}
}

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
	r._unLockOnce.Do(r.initUnLockScript)
	prefix, _, _ := parseLockOptions(options)
	key = fmt.Sprintf("%s:%s", prefix, key)
	result, err := r._unLockScript.Run(ctx, r, []string{key}, []string{identifier}).Int()
	if err != nil {
		return err
	}
	if result == 1 {
		return nil
	}
	return fmt.Errorf("unlock failed: key %s is no longer valid or mismatched identifier", key)
}

func (r *Redis) RenewLock(ctx context.Context, key, identifier string, expiration time.Duration, options ...Option) (err error) {
	r._renewLockOnce.Do(r.initRenewLockScript)
	prefix, _, _ := parseLockOptions(options)
	key = fmt.Sprintf("%s:%s", prefix, key)
	result, err := r._renewLockScript.Run(ctx, r, []string{key}, identifier, expiration.Milliseconds()).Int()
	if err != nil {
		return err
	}
	if result == 1 {
		return nil
	}
	return fmt.Errorf("renew failed: key %s is no longer valid or mismatched identifier", key)
}

// IncrWithExpire IncrWithExpire: 如果是第一次，则会添加过期时间
func (r *Redis) IncrWithExpire(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	r._incrWithExpireLockOnce.Do(r.initIncrWithExpireScript)
	result, err := r._incrWithExpireScript.Run(ctx, r, []string{key}, expiration.Milliseconds()).Int64()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *Redis) AllowFixedLimit(ctx context.Context, key string, expiration time.Duration, maxCount uint32) (bool, error) {
	r._allowFixedLimitOnce.Do(r.initAllowFixedLimitScript)
	result, err := r._allowFixedLimitScript.Run(ctx, r, []string{key}, expiration.Milliseconds(), maxCount).Int64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
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

func (r *Redis) initUnLockScript() {
	r._unLockScript = redis.NewScript(unLockLua)
}

func (r *Redis) initRenewLockScript() {
	r._renewLockScript = redis.NewScript(renewLockLua)
}

func (r *Redis) initIncrWithExpireScript() {
	r._incrWithExpireScript = redis.NewScript(incrWithExpireLua)
}

func (r *Redis) initAllowFixedLimitScript() {
	r._allowFixedLimitScript = redis.NewScript(fixedWindowLimitLua)
}
