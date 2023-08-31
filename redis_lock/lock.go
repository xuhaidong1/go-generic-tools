package redis_lock

import (
	"context"
	_ "embed"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/xuhaidong1/go-generic-tools/redis_lock/errs"
	"sync"
	"time"
)

var (
	//go:embed lua/unlock.lua
	luaUnlock string
	//go:embed lua/lock.lua
	luaLock string
	//go:embed lua/refresh.lua
	luaRefresh string
)

type RetryStrategy interface {
	Next() (time.Duration, bool)
}

type Client struct {
	client redis.Cmdable
}

type Lock struct {
	client     redis.Cmdable
	key        string
	value      string
	expiration time.Duration
	unlock     chan struct{}
	unlockOnce sync.Once
}

func (c *Client) Lock(ctx context.Context, key string, expiration time.Duration, timeout time.Duration, retry RetryStrategy) (*Lock, error) {
	val := uuid.New().String()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	for {
		lctx, cancel := context.WithTimeout(ctx, timeout)
		res, err := c.client.Eval(lctx, luaLock, []string{key}, val, expiration.Seconds()).Result()
		cancel()
		if res == "OK" {
			return &Lock{
				client:     c.client,
				key:        key,
				value:      val,
				expiration: expiration,
				unlock:     make(chan struct{}, 1),
			}, nil
		}
		if err != nil && err != context.DeadlineExceeded {
			return nil, err
		}
		//超时 或者锁被别人拿着了
		interval, ok := retry.Next()
		if !ok {
			return nil, errs.NewErrFailedToPreemptLock()
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):

		}
	}

}

func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	val := uuid.New().String()
	ok, err := c.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errs.NewErrFailedToPreemptLock()
	}
	return &Lock{
		client:     c.client,
		key:        key,
		value:      val,
		expiration: expiration,
		unlock:     make(chan struct{}, 1),
	}, nil
}

func (l *Lock) Unlock(ctx context.Context) error {
	l.unlockOnce.Do(func() {
		close(l.unlock)
	})
	//有锁就且是我的就删掉，没锁就算了，要考虑，用 lua 脚本来封装检查-删除的两个步骤
	res, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.value).Int64()
	if err != nil {
		return err
	}
	if res != 1 {
		return errs.NewErrLockNotHold()
	}
	return nil
}

func (l *Lock) Refresh(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaRefresh, []string{l.key}, l.value, l.expiration.Seconds()).Int64()
	//if err == redis.Nil {
	//	return errs.NewErrLockNotHold()
	//}
	if err != nil {
		return err
	}
	if res != 1 {
		return errs.NewErrLockNotHold()
	}
	return nil
}

func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	retrySignal := make(chan struct{}, 1)
	//不断续约 直到收到退出信号
	defer close(retrySignal)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err == context.DeadlineExceeded {
				retrySignal <- struct{}{}
				continue
			}
			if err != nil {
				//不可挽回的错误 要考虑中断业务执行
				return err
			}
		case <-retrySignal:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err == context.DeadlineExceeded {
				//可以超时重试
				retrySignal <- struct{}{}
				continue
			}
			if err != nil {
				//不可挽回的错误 要考虑中断业务执行
				return err
			}
		case <-l.unlock:
			//解锁的时候停止续约
			return nil
		}

	}

}
