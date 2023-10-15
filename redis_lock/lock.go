package redis_lock

import (
	"context"
	_ "embed"
	"fmt"
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

// RetryStrategy 加锁重试策略 next返回的是要过多久进行下一次抢锁，bool为false是不抢锁了
type RetryStrategy interface {
	Next() (time.Duration, bool)
}

type Client struct {
	client redis.Cmdable
	retry  RetryStrategy
}

func NewClient(client redis.Cmdable, opts ...Options) *Client {
	c := &Client{client: client}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Options func(c *Client)

func WithRetryStrategy(r RetryStrategy) Options {
	return func(c *Client) {
		c.retry = r
	}
}

type Lock struct {
	client     redis.Cmdable
	key        string
	value      string
	expiration time.Duration
	unlock     chan struct{}
	unlockOnce sync.Once
}

// Lock 加锁重试，可以注入重试策略
// timeout：抢锁的超时时间
func (c *Client) Lock(ctx context.Context, key, val string, expiration, timeout time.Duration) (*Lock, error) {
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
		//有err但不是超时
		if err != nil && err != context.DeadlineExceeded {
			return nil, err
		}
		//没有重试策略就直接返回错误
		if c.retry == nil {
			return nil, err
		}
		interval, ok := c.retry.Next()
		//超时 或者锁被别人拿着了
		if !ok {
			if err != nil {
				err = fmt.Errorf("最后一次重试错误:%w", err)
			} else {
				err = fmt.Errorf("锁被人持有:%w", errs.NewErrFailedToPreemptLock())
			}
			return nil, fmt.Errorf("重试机会耗尽,%w", err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}

}

func (c *Client) TryLock(ctx context.Context, key, val string, expiration time.Duration) (*Lock, error) {
	res, err := c.client.Eval(ctx, luaLock, []string{key}, val, expiration.Seconds()).Result()
	if err != nil {
		return nil, errs.NewErrFailedToPreemptLock()
	}
	if res == "OK" {
		return &Lock{
			client:     c.client,
			key:        key,
			value:      val,
			expiration: expiration,
			unlock:     make(chan struct{}, 1),
		}, nil
	}
	return nil, errs.NewErrLockNotHold()
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
