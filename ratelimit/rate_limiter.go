package ratelimit

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	r *rate.Limiter
}

// NewLimiter 创建一个限流器，允许在 duration 内执行 maxRequests 次请求
// 例如：NewLimiter(5*time.Second, 10, 5) 表示每5秒最多10次请求，突发容量5
func NewLimiter(duration time.Duration, maxRequests, burst uint64) *Limiter {
	interval := float64(duration) / float64(maxRequests)
	rl := rate.NewLimiter(
		rate.Every(time.Duration(interval)),
		int(burst),
	)
	return &Limiter{r: rl}
}

// Allow 检查是否允许请求
func (l *Limiter) Allow() bool {
	return l.r.Allow()
}

// Wait 等待直到可以执行请求
func (l *Limiter) Wait(ctx context.Context) error {
	return l.r.Wait(ctx)
}
