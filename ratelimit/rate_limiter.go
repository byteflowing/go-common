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
	var rl *rate.Limiter
	if maxRequests == 0 {
		rl = rate.NewLimiter(rate.Limit(0), int(burst)) // 禁止所有请求
	} else {
		interval := duration / time.Duration(maxRequests)
		rl = rate.NewLimiter(
			rate.Every(interval),
			int(burst),
		)
	}
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

func (l *Limiter) Tokens() float64 {
	return l.r.Tokens()
}

func (l *Limiter) Burst() int {
	return l.r.Burst()
}

func (l *Limiter) SetLimit(duration time.Duration, maxRequests uint64) {
	if maxRequests == 0 {
		l.r.SetLimit(rate.Limit(0)) // 禁止所有请求
		return
	}
	interval := duration / time.Duration(maxRequests)
	newLimit := rate.Every(interval)
	l.r.SetLimit(newLimit)
}

func (l *Limiter) SetBurst(burst uint64) {
	l.r.SetBurst(int(burst))
}
