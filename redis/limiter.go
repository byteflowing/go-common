package redis

import (
	"context"
	"fmt"
	"time"

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
	Tag      string        // 标签，方便业务方感知被限流的窗口
}

type Limiter struct {
	rdb     *Redis
	prefix  string
	windows []*Window
	script  *redis.Script
}

// NewLimiter : 创建带有滑动窗口的redis限流器，例如短信发送限制 1min 1次 5min 3次 1天 10次
// @param prefix 为key的前缀 e.g. prefix: "sms:limit"
// @param windows参数需要按照依次递增的顺序传递进来
// @param window.Duration 最小只能支持到秒级
// @param window.Tag 用于业务标记方便前端拼接弹窗信息
func NewLimiter(rdb *Redis, prefix string, windows []*Window) *Limiter {
	return &Limiter{
		rdb:     rdb,
		prefix:  prefix,
		windows: windows,
		script:  getSlidingWindowScript(),
	}
}

// Allow 判断key是否被允许
// @return allowed 是否被允许
// @return blocked 如果不允许返回被限制的窗口
// @return retryAfter 如果被限制还有多少s会解除限制
// @return err 错误信息
func (l *Limiter) Allow(ctx context.Context, key string) (allowed bool, blocked *Window, retryAfter int64, err error) {
	var redisKeys []string
	var args []interface{}

	for _, win := range l.windows {
		redisKey := fmt.Sprintf("%s:{%s}:%ds", l.prefix, key, int(win.Duration.Seconds()))
		redisKeys = append(redisKeys, redisKey)
		args = append(args, int(win.Duration.Seconds()), win.Limit)
	}

	res, err := l.script.Run(ctx, l.rdb, redisKeys, args...).Result()
	if err != nil {
		return false, nil, 0, err
	}

	arr, ok := res.([]interface{})
	if !ok || len(arr) != 2 {
		return false, nil, 0, fmt.Errorf("unexpected Lua result: %#v", res)
	}

	code, _ := arr[0].(int64)
	ttl, _ := arr[1].(int64)

	if code == 1 {
		return true, nil, 0, nil
	}

	idx := code - 100
	if idx < 0 || int(idx) >= len(l.windows) {
		return false, nil, 0, fmt.Errorf("invalid limit index %d", idx)
	}

	return false, l.windows[idx], ttl, nil
}

func (l *Limiter) Reset(ctx context.Context, key string) error {
	var redisKeys []string
	for _, win := range l.windows {
		redisKey := fmt.Sprintf("%s:{%s}:%ds", l.prefix, key, int(win.Duration.Seconds()))
		redisKeys = append(redisKeys, redisKey)
	}
	return l.rdb.Del(ctx, redisKeys...).Err()
}
