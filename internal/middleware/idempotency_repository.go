package middleware

import (
	"context"
	"github.com/go-redis/redis/v8"
	"sync"
	"time"
)

type Status int

const (
	Unknown Status = iota
	Finished
)

type IdempotencyKeysRepository interface {
	Status(key string) (Status, error)
	SetStatus(key string, status Status) error
}

type InMemIdempotencyKeysRepository struct {
	m *sync.Map
}

func NewInMemIdempotencyKeysRepository() IdempotencyKeysRepository {
	return &InMemIdempotencyKeysRepository{
		m: new(sync.Map),
	}
}

func (r *InMemIdempotencyKeysRepository) Status(key string) (Status, error) {
	s, ok := r.m.Load(key)
	if !ok {
		return Unknown, nil
	}
	return s.(Status), nil
}

func (r *InMemIdempotencyKeysRepository) SetStatus(key string, status Status) error {
	r.m.Store(key, status)
	return nil
}

type RedisRepository struct {
	db *redis.Client
}

func NewRedisIdempotencyKeysRepository(addr, password string) IdempotencyKeysRepository {
	return &RedisRepository{db: redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})}
}

func (r *RedisRepository) Status(key string) (Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	s, err := r.db.Get(ctx, key).Int()
	if err == redis.Nil {
		return Unknown, nil
	}

	if err != nil {
		return Unknown, err
	}

	return Status(s), nil
}

func (r *RedisRepository) SetStatus(key string, status Status) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return r.db.Set(ctx, key, int(status), 0).Err()
}
