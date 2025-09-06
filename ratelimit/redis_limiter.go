package ratelimit

import (
	"context"
	"fmt"
	"time"

	redisWrapper "github.com/byteflowing/go-common/redis"
	"github.com/byteflowing/go-common/syncx"
	"github.com/byteflowing/go-common/trans"
	limiterv1 "github.com/byteflowing/proto/gen/go/limiter/v1"
	"github.com/redis/go-redis/v9"
)

const slidingWindowLua = `
for i = 1, #KEYS do
	local key = KEYS[i]
	local duration = tonumber(ARGV[(i - 1) * 2 + 1])
	local limit = tonumber(ARGV[(i - 1) * 2 + 2])
	local cnt = redis.call("INCR", key)
	if cnt == 1 then
		redis.call("EXPIRE", key, duration)
	end
	if cnt > limit then
		local ttl = redis.call("TTL", key)
		if ttl >= 0 then
			return {100 + (i - 1), ttl}
		end
	end
end
return {1, 0}
`

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

type RedisLimiter struct {
	rdb     *redisWrapper.Redis
	prefix  string
	windows []*limiterv1.LimitRule
	script  *redis.Script
}

// NewRedisLimiter : 创建带有滑动窗口的redis限流器，例如短信发送限制 1min 1次 5min 3次 1天 10次
// @param prefix 为key的前缀 e.g. prefix: "sms:limit"
// @param windows参数需要按照依次递增的顺序传递进来
// @param window.Duration 最小只能支持到秒级
// @param window.Tag 用于业务标记方便前端拼接弹窗信息
func NewRedisLimiter(rdb *redisWrapper.Redis, prefix string, rules []*limiterv1.LimitRule) *RedisLimiter {
	return &RedisLimiter{
		rdb:     rdb,
		prefix:  prefix,
		windows: rules,
		script:  getSlidingWindowScript(),
	}
}

// Allow 判断key是否被允许
// @return allowed 是否被允许
// @return blocked 如果不允许返回被限制的窗口
// @return retryAfter 如果被限制还有多少s会解除限制
// @return err 错误信息
func (l *RedisLimiter) Allow(ctx context.Context, key string) (allowed bool, rule *limiterv1.LimitRule, err error) {
	var redisKeys []string
	var args []interface{}

	for _, win := range l.windows {
		redisKey := l.getRedisKey(key, int(win.Duration.Seconds))
		redisKeys = append(redisKeys, redisKey)
		args = append(args, int(win.Duration.Seconds), win.Limit)
	}

	res, err := l.script.Run(ctx, l.rdb, redisKeys, args...).Result()
	if err != nil {
		return false, nil, err
	}

	arr, ok := res.([]interface{})
	if !ok || len(arr) != 2 {
		return false, nil, fmt.Errorf("unexpected Lua result: %#v", res)
	}
	code, _ := arr[0].(int64)
	ttl, _ := arr[1].(int64)

	if code == 1 {
		return true, nil, nil
	}

	idx := code - 100
	if idx < 0 || int(idx) >= len(l.windows) {
		return false, nil, fmt.Errorf("invalid limit index %d", idx)
	}
	rule = l.windows[idx]
	rule.RetryAfter = trans.Int64(ttl)
	return false, rule, nil
}

func (l *RedisLimiter) Reset(ctx context.Context, key string) error {
	var redisKeys []string
	for _, win := range l.windows {
		redisKey := l.getRedisKey(key, int(win.Duration.Seconds))
		redisKeys = append(redisKeys, redisKey)
	}
	return l.rdb.Del(ctx, redisKeys...).Err()
}

func (l *RedisLimiter) getRedisKey(key string, win int) string {
	return fmt.Sprintf("%s:{%s}:%ds", l.prefix, key, win)
}
