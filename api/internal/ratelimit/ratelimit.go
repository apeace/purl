package ratelimit

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// script atomically records a request within a sliding window and reports whether it is
// within the configured limit. Returns {1, 0} if allowed; {0, retry_after_ms} if exceeded.
var script = redis.NewScript(`
local key = KEYS[1]
local now_ms = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local max = tonumber(ARGV[3])
local member = ARGV[4]

redis.call('ZREMRANGEBYSCORE', key, '-inf', now_ms - window_ms)
local count = redis.call('ZCARD', key)
if count < max then
    redis.call('ZADD', key, now_ms, member)
    redis.call('PEXPIRE', key, window_ms + 1000)
    return {1, 0}
end
local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
if #oldest > 0 then
    local wait_ms = math.ceil((tonumber(oldest[2]) + window_ms) - now_ms)
    if wait_ms < 1 then wait_ms = 1 end
    return {0, wait_ms}
end
return {0, window_ms}
`)

// Limiter is a Redis-backed sliding-window rate limiter.
type Limiter struct {
	rdb    *redis.Client
	name   string
	max    int64
	window time.Duration
}

var seq atomic.Int64

// New creates a Limiter for the named resource, allowing at most max requests per window.
func New(rdb *redis.Client, name string, max int64, window time.Duration) *Limiter {
	return &Limiter{rdb: rdb, name: name, max: max, window: window}
}

// Wait blocks until the rate limit allows a request for the given key, then records it.
// Pass key="" for a global (keyless) limit. Returns an error if ctx is cancelled.
func (l *Limiter) Wait(ctx context.Context, key string) error {
	for {
		allowed, retryAfter, err := l.tryAcquire(ctx, key)
		if err != nil {
			return err
		}
		if allowed {
			return nil
		}
		if retryAfter <= 0 {
			retryAfter = time.Millisecond
		}
		log.Printf("rate limit %s: waiting %s", l.name, retryAfter.Round(time.Millisecond))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryAfter):
		}
	}
}

// tryAcquire attempts to consume one slot in the sliding window.
// Returns (true, 0, nil) on success, or (false, retryAfter, nil) if the limit is exceeded.
func (l *Limiter) tryAcquire(ctx context.Context, key string) (bool, time.Duration, error) {
	nowMs := time.Now().UnixMilli()
	member := fmt.Sprintf("%d-%d", nowMs, seq.Add(1))

	res, err := script.Run(ctx, l.rdb,
		[]string{l.redisKey(key)},
		nowMs, l.window.Milliseconds(), l.max, member,
	).Slice()
	if err != nil {
		return false, 0, fmt.Errorf("rate limit script: %w", err)
	}

	allowed := res[0].(int64) == 1
	retryMs := res[1].(int64)
	return allowed, time.Duration(retryMs) * time.Millisecond, nil
}

func (l *Limiter) redisKey(key string) string {
	if key == "" {
		return fmt.Sprintf("ratelimit:%s:_global_", l.name)
	}
	return fmt.Sprintf("ratelimit:%s:%s", l.name, key)
}
