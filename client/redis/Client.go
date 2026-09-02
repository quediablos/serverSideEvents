package redis

import (
	"context"
	"fmt"
	"serverSideEvents/candle"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func New(addr, password string, db int) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Client{client: rdb}
}

func (r *Client) Get(ctx context.Context, key string, field string) (string, error) {
	val, err := r.client.HGet(ctx, key, field).Result()

	if err == redis.Nil {
		return "", fmt.Errorf("key %q not found", key)
	}
	if err != nil {
		return "", fmt.Errorf("redis get error: %w", err)
	}
	return val, nil
}

func (r *Client) HSet(ctx context.Context, key string, field string, value string) error {
	err := r.client.HSet(ctx, key, field, value).Err()
	if err != nil {
		return fmt.Errorf("redis hset error: %w", err)
	}
	return nil
}

func (r *Client) GetCandle(ctx context.Context, key string) (candle.Candle, error) {
	vals, err := r.client.HMGet(ctx, key, "open", "high", "low", "close").Result()
	if err != nil {
		return candle.Candle{}, fmt.Errorf("redis hmget error: %w", err)
	}

	toString := func(v any) string {
		if v == nil {
			return ""
		}
		return v.(string)
	}

	return candle.Candle{
		Open:  toString(vals[0]),
		High:  toString(vals[1]),
		Low:   toString(vals[2]),
		Close: toString(vals[3]),
	}, nil
}

func (r *Client) GetMany(ctx context.Context, keys []string, field string) ([]string, error) {
	pipe := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.HGet(ctx, key, field)
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("redis pipeline error: %w", err)
	}

	results := make([]string, len(keys))
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			results[i] = ""
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("redis get error for key %q: %w", keys[i], err)
		}
		results[i] = val
	}
	return results, nil
}

func (r *Client) HSetMultiple(ctx context.Context, key string, values ...any) error {
	err := r.client.HSet(ctx, key, values).Err()
	if err != nil {
		return fmt.Errorf("redis hset error: %w", err)
	}
	return nil
}
