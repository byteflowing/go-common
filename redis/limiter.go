package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/byteflowing/go-common/syncx"
	"github.com/redis/go-redis/v9"
)

var script *redis.Script
var initScript = syncx.Once(func() {
	script = redis.NewScript(slidingWindowLua)
})

func getSlidingWindowScript() *redis.Script {
	initScript()
	return script
}

type Window struct {
	Duration time.Duration // 限制周期
	Limit    uint32        // 最大次数
}

type Limiter struct {
	rdb     *Redis
	prefix  string
	windows []*Window
	ttl     int
	script  *redis.Script
}

// NewLimiter : prefix为key的前缀
// e.g. prefix: "sms:limit"
func NewLimiter(rdb *Redis, prefix string, windows []*Window) *Limiter {
	longest := 0
	for _, window := range windows {
		if int(window.Duration.Seconds()) > longest {
			longest = int(window.Duration.Seconds())
		}
	}
	return &Limiter{
		rdb:     rdb,
		prefix:  prefix,
		windows: windows,
		ttl:     longest,
		script:  getSlidingWindowScript(),
	}
}

// Allow : 如果没被限流则返回true, nil,nil
// 如果被限流则返回 false, Window为被限流的窗口
func (l *Limiter) Allow(ctx context.Context, key string) (bool, *Window, error) {
	redisKey := fmt.Sprintf("%s:{%s}", l.prefix, key)
	now := time.Now().Unix()
	id := fmt.Sprintf("%d-%d", now, fastrand.Intn(100000))
	args := l.buildArgs(now, id)
	res, err := l.script.Run(ctx, l.rdb, []string{redisKey}, args...).Result()
	if err != nil {
		return false, nil, err
	}
	num, ok := res.(int64)
	if !ok {
		return false, nil, fmt.Errorf("unexpected result type %T", res)
	}
	// 1 表示没有被限流
	if num == 1 {
		return true, nil, nil
	}
	// 被限流返回对应窗口
	if num >= 100 {
		index := int(num - 100)
		return false, l.windows[index], nil
	}
	return false, nil, fmt.Errorf("unexpected return value: %d", num)
}

func (l *Limiter) buildArgs(now int64, id string) []interface{} {
	args := []interface{}{now, id, l.ttl}
	for _, window := range l.windows {
		args = append(args, int(window.Duration.Seconds()), window.Limit)
	}
	return args
}
