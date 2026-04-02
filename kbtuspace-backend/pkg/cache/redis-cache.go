package cache

import (
	"context"
	"errors"
	"time"

	"kbtuspace-backend/internal/models"

	json "github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client  *redis.Client
	expires time.Duration
}

func NewRedisCache(redisURL string, exp time.Duration) (PostsCache, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &redisCache{client: client, expires: exp}, nil
}

func (cache *redisCache) SetPost(key string, value *models.Post) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return cache.client.Set(context.Background(), key, payload, cache.expires).Err()
}

func (cache *redisCache) GetPost(key string) (*models.Post, bool, error) {
	data, err := cache.client.Get(context.Background(), key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var post models.Post
	if err := json.Unmarshal(data, &post); err != nil {
		return nil, false, err
	}

	return &post, true, nil
}

func (cache *redisCache) SetPosts(key string, value []models.Post) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return cache.client.Set(context.Background(), key, payload, cache.expires).Err()
}

func (cache *redisCache) GetPosts(key string) ([]models.Post, bool, error) {
	data, err := cache.client.Get(context.Background(), key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var posts []models.Post
	if err := json.Unmarshal(data, &posts); err != nil {
		return nil, false, err
	}

	return posts, true, nil
}

func (cache *redisCache) Delete(key string) error {
	return cache.client.Del(context.Background(), key).Err()
}

func (cache *redisCache) DeletePrefix(prefix string) error {
	ctx := context.Background()
	iter := cache.client.Scan(ctx, 0, prefix+"*", 0).Iterator()
	keys := make([]string, 0)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}

	return cache.client.Del(ctx, keys...).Err()
}
