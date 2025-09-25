package cache

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct{ R *redis.Client }

func New() *Store {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return nil
	}
	r := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	return &Store{R: r}
}

func (s *Store) Close() {
	if s != nil && s.R != nil {
		_ = s.R.Close()
	}
}

func (s *Store) GetJSON(ctx context.Context, key string, dst any) (bool, error) {
	if s == nil {
		return false, nil
	}
	b, err := s.R.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(b, dst)
}

func (s *Store) SetJSON(ctx context.Context, key string, val any, ttl time.Duration) error {
	if s == nil {
		return nil
	}
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return s.R.Set(ctx, key, b, ttl).Err()
}

func (s *Store) DelPattern(ctx context.Context, pat string) {
	if s == nil {
		return
	}
	iter := s.R.Scan(ctx, 0, pat, 100).Iterator()
	for iter.Next(ctx) {
		_ = s.R.Del(ctx, iter.Val()).Err()
	}
	_ = iter.Err()
}

func (s *Store) RevokeJTI(ctx context.Context, jti string, ttl time.Duration) error {
	if s == nil {
		return nil
	}
	return s.R.Set(ctx, "rev:"+jti, "1", ttl).Err()
}
func (s *Store) IsRevoked(ctx context.Context, jti string) (bool, error) {
	if s == nil {
		return false, nil
	}
	_, err := s.R.Get(ctx, "rev:"+jti).Result()
	if err == redis.Nil {
		return false, nil
	}
	return err == nil, err
}
