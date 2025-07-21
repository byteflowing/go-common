package redis

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getTestRedis() *Redis {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	cfg := &Config{
		Addr:     addr,
		DB:       15, // 用测试库，避免污染
		Protocol: 2,
	}
	rdb := New(cfg)
	_ = rdb.FlushDB(context.Background()).Err() // 清空测试库
	return rdb
}

func TestLimiter_Allow_Basic(t *testing.T) {
	rdb := getTestRedis()
	limiter := NewLimiter(rdb, "test:limit", []*Window{{Duration: 2 * time.Second, Limit: 2}})
	ctx := context.Background()
	key := "user1"

	// 第一次允许
	ok, win, err := limiter.Allow(ctx, key)
	assert.True(t, ok)
	assert.Nil(t, win)
	assert.NoError(t, err)

	// 第二次允许
	ok, win, err = limiter.Allow(ctx, key)
	assert.True(t, ok)
	assert.Nil(t, win)
	assert.NoError(t, err)

	// 第三次被限流
	ok, win, err = limiter.Allow(ctx, key)
	assert.False(t, ok)
	assert.NotNil(t, win)
	assert.NoError(t, err)
	assert.Equal(t, uint32(2), win.Limit)

	// 等待窗口过期
	time.Sleep(2 * time.Second)
	ok, win, err = limiter.Allow(ctx, key)
	assert.True(t, ok)
	assert.Nil(t, win)
	assert.NoError(t, err)
}

func TestLimiter_Allow_MultiWindow(t *testing.T) {
	rdb := getTestRedis()
	limiter := NewLimiter(rdb, "test:multi", []*Window{
		{Duration: 2 * time.Second, Limit: 2},
		{Duration: 10 * time.Second, Limit: 3},
	})
	ctx := context.Background()
	key := "user2"

	// 2秒2次，5秒3次
	for i := 0; i < 2; i++ {
		ok, _, err := limiter.Allow(ctx, key)
		assert.True(t, ok)
		assert.NoError(t, err)
	}
	// 2秒窗口被限流
	ok, win, err := limiter.Allow(ctx, key)
	assert.False(t, ok)
	assert.Equal(t, uint32(2), win.Limit)
	assert.NoError(t, err)

	// 等2秒，2秒窗口恢复，但5秒窗口还没恢复
	time.Sleep(5 * time.Second)
	ok, _, err = limiter.Allow(ctx, key)
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, win, err = limiter.Allow(ctx, key)
	assert.False(t, ok)
	assert.Equal(t, uint32(3), win.Limit)
	assert.NoError(t, err)
}
