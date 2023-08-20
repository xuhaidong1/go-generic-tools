package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisCache struct {
	client redis.Cmdable //cmdable方便使用gomock
}

// NewRedisCache 面向接口编程，依赖注入，不要传一个string的地址自己建redisClient，要不然单元测试就会尝试连这个addr，没办法测，我们需要mockredis
func NewRedisCache(client redis.Cmdable) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Get(ctx context.Context, key string) (any, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	_, err := r.client.Set(ctx, key, val, expiration).Result() //成功了redis会返回一个OK
	return err
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	_, err := r.client.Del(ctx, key).Result()
	return err
}
