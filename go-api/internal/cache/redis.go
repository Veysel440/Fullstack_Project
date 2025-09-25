package cache

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func New() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return nil
	}
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
	})
}

func Ping(ctx context.Context, rdb *redis.Client) error {
	if rdb == nil {
		return nil
	}
	_, err := rdb.Ping(ctx).Result()
	return err
}

var luaTokenBucket = `
local key     = KEYS[1]
local now     = tonumber(ARGV[1])
local rate    = tonumber(ARGV[2])
local burst   = tonumber(ARGV[3])
local period  = tonumber(ARGV[4]) -- ms

local data = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(data[1]) or burst
local ts     = tonumber(data[2]) or now

local delta = (now - ts) / period * rate
tokens = math.min(burst, tokens + delta)
if tokens < 1 then
  redis.call("HMSET", key, "tokens", tokens, "ts", now)
  redis.call("PEXPIRE", key, 60000)
  return 0
else
  tokens = tokens - 1
  redis.call("HMSET", key, "tokens", tokens, "ts", now)
  redis.call("PEXPIRE", key, 60000)
  return 1
end
`

type Limiter struct {
	RDB   *redis.Client
	Rate  float64
	Burst float64
}

func (l *Limiter) Allow(ctx context.Context, key string) bool {
	if l.RDB == nil {
		return true
	}
	now := float64(time.Now().UnixMilli())
	n, err := l.RDB.Eval(ctx, luaTokenBucket, []string{"rl:" + key},
		now, l.Rate, l.Burst, float64(1000)).Int()
	return err == nil && n == 1
}
