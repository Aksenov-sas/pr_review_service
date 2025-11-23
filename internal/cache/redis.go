package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"

	"pr-reviewer-assignment-service/internal/interfaces"
)

// RedisCache реализация кэша на Redis
type RedisCache struct {
	client *redis.Client
	logger *slog.Logger
}

// NewRedisCache создает новый экземпляр Redis кэша
func NewRedisCache(redisURL string, logger *slog.Logger) (*RedisCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Проверим соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ошибка пинга Redis: %w", err)
	}

	cache := &RedisCache{
		client: client,
		logger: logger,
	}

	return cache, nil
}

// Close закрывает соединение с Redis
func (c *RedisCache) Close() error {
	return c.client.Close()
}

var _ interfaces.Cache = (*RedisCache)(nil) // Проверяем, что RedisCache реализует интерфейс Cache

// Get получает значение из кэша
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("ключ не найден: %w", err)
		}
		return fmt.Errorf("ошибка получения из Redis: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("ошибка десериализации данных: %w", err)
	}

	return nil
}

// Set устанавливает значение в кэше
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	if err := c.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("ошибка сохранения в Redis: %w", err)
	}

	return nil
}

// Delete удаляет значение из кэша
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("ошибка удаления из Redis: %w", err)
	}

	return nil
}

// Exists проверяет существование ключа в кэше
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("ошибка проверки существования ключа: %w", err)
	}

	return count > 0, nil
}
