package authbff

import (
	"context"
	"encoding/json"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type RedisStore struct {
	cli *redis.Client
	ttl time.Duration
}

func NewRedisStore(addr string, ttl time.Duration) (*RedisStore, error) {
	cli := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &RedisStore{cli: cli, ttl: ttl}, nil
}

func (s *RedisStore) Set(sess *Session) {
	b, _ := json.Marshal(sess)
	ctx := context.Background()
	s.cli.Set(ctx, s.key(sess.ID), b, time.Until(sess.ExpiresAt))
}

func (s *RedisStore) Get(id string) (*Session, bool) {
	ctx := context.Background()
	b, err := s.cli.Get(ctx, s.key(id)).Bytes()
	if err != nil {
		return nil, false
	}
	var sess Session
	if err := json.Unmarshal(b, &sess); err != nil {
		return nil, false
	}
	return &sess, true
}

func (s *RedisStore) Delete(id string) {
	ctx := context.Background()
	_ = s.cli.Del(ctx, s.key(id)).Err()
}

func (s *RedisStore) key(id string) string { return "auth:sid:" + id }
