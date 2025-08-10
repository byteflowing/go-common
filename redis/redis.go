package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/byteflowing/go-common/idx"
	"github.com/byteflowing/go-common/syncx"
	"github.com/byteflowing/go-common/timex"
	"github.com/redis/go-redis/v9"
)

const (
	Nil = redis.Nil
)

var (
	ErrLock                = errors.New("lock failed")
	ErrLockAlreadyReleased = errors.New("lock already released")
)

const (
	defaultKeyPrefix    = "lock"
	defaultTries        = 5
	defaultWaitDuration = 5 * time.Millisecond
	randMills           = 120000
)

type Options struct {
	LockKeyPrefix    string
	LockTryTimes     int
	LockWaitDuration time.Duration
}

type Option func(o *Options)

type Redis struct {
	redis.Cmdable
	initUnLockScript          func()
	initRenewLockScript       func()
	initAllowFixedLimitScript func()
	initIncrWithExpireScript  func()
	initDailyLimitScript      func()
	unLockScript              *redis.Script
	renewLockScript           *redis.Script
	allowFixedLimitScript     *redis.Script
	incrWithExpireScript      *redis.Script
	dailyLimitScript          *redis.Script
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
	r := &Redis{Cmdable: rdb}
	r.initUnLockScript = syncx.Once(func() { r.unLockScript = redis.NewScript(unLockLua) })
	r.initRenewLockScript = syncx.Once(func() { r.renewLockScript = redis.NewScript(renewLockLua) })
	r.initIncrWithExpireScript = syncx.Once(func() { r.incrWithExpireScript = redis.NewScript(incrWithExpireLua) })
	r.initAllowFixedLimitScript = syncx.Once(func() { r.allowFixedLimitScript = redis.NewScript(allowFixedLimitLua) })
	r.initDailyLimitScript = syncx.Once(func() { r.dailyLimitScript = redis.NewScript(allowDailyLimitLua) })
	return r
}

func (r *Redis) Lock(ctx context.Context, key string, expiration time.Duration, options ...Option) (identifier string, err error) {
	identifier = idx.UUIDv4()
	waitDuration, tries := parseLockOptions(options)
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
	return "", ErrLock
}

func (r *Redis) Unlock(ctx context.Context, key, identifier string, options ...Option) (err error) {
	r.initUnLockScript()
	result, err := r.unLockScript.Run(ctx, r, []string{key}, []string{identifier}).Int()
	if err != nil {
		return err
	}
	if result == 1 {
		return nil
	}
	return ErrLockAlreadyReleased
}

func (r *Redis) RenewLock(ctx context.Context, key, identifier string, expiration time.Duration, options ...Option) (err error) {
	r.initRenewLockScript()
	result, err := r.renewLockScript.Run(ctx, r, []string{key}, identifier, expiration.Milliseconds()).Int()
	if err != nil {
		return err
	}
	if result == 1 {
		return nil
	}
	return ErrLockAlreadyReleased
}

// IncrWithExpire IncrWithExpire: 如果是第一次，则会添加过期时间
func (r *Redis) IncrWithExpire(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	r.initIncrWithExpireScript()
	result, err := r.incrWithExpireScript.Run(ctx, r, []string{key}, expiration.Milliseconds()).Int64()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// AllowFixedLimit 用于duration时间内的限流次数
func (r *Redis) AllowFixedLimit(ctx context.Context, key string, expiration time.Duration, maxCount uint32) (bool, error) {
	r.initAllowFixedLimitScript()
	result, err := r.allowFixedLimitScript.Run(ctx, r, []string{key}, expiration.Milliseconds(), maxCount).Int64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

// AllowDailyLimit 用于当天的限流次数
// 常用于按自然天算api请求次数场景
func (r *Redis) AllowDailyLimit(ctx context.Context, prefix, target string, maxCount uint32) (bool, error) {
	r.initDailyLimitScript()
	ttl := timex.EndOfDayMillis()
	// 加上随机值，避免0点大量key同时过期
	randTTL := ttl + fastrand.Int63n(randMills)
	todayKey := time.UnixMilli(ttl).Format("20060102")
	key := fmt.Sprintf("%s:%s:%s", prefix, todayKey, target)
	result, err := r.allowFixedLimitScript.Run(ctx, r, []string{key}, randTTL, maxCount).Int64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
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

func parseLockOptions(options []Option) (waitDuration time.Duration, tries int) {
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
	return
}
