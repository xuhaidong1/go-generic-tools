package cache

import (
	"context"
	"time"
)

// Cache 值的类型问题：
// string:可以，问题是本地缓存，结构体转化为string，用json序列化
// []byte最通用的表达，可以存储序列化后的数据，也可以存加密数据，但是用起来不方便
// any：方便，redis之类的实现，需要考虑序列化问题
// 提高内存使用率：不要直接存对象，序列化（紧凑的算法如protobuf）+压缩
// 不使用map，用树形结构等
// 为什么不定义泛型？定义泛型用户只能在cache里面存一种数据了
type Cache interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, val any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

type LoadFunc func(ctx context.Context, key string) (any, error)
type StoreFunc func(ctx context.Context, key string, val any) error
type BloomFilter func(ctx context.Context, key string) bool
