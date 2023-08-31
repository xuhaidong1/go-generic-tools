package redis_lock

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/xuhaidong1/go-generic-tools/redis_lock/errs"
	"testing"
	"time"
)

type ClientE2ESuite struct {
	suite.Suite
	rdb redis.Cmdable
}

func (s *ClientE2ESuite) SetupSuite() {
	s.rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// 确保测试的目标 Redis 已经启动成功了
	for s.rdb.Ping(context.Background()).Err() != nil {

	}
}

func TestClientE2E(t *testing.T) {
	suite.Run(t, &ClientE2ESuite{})
}

func ExampleLock_Refresh() {

	var l *Lock
	ch := make(chan struct{}, 1)
	stop := make(chan struct{})
	//bizStop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		//不断续约 直到收到退出信号
		defer ticker.Stop()
		retryCnt := 0
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				err := l.Refresh(ctx)
				cancel()

				if err == context.DeadlineExceeded {
					//可以超时重试
					ch <- struct{}{}
					continue
				}
				if err != nil {
					//不可挽回的错误 要考虑中断业务执行
					return
				}
				retryCnt = 0
			case <-ch:
				retryCnt++
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				err := l.Refresh(ctx)
				cancel()
				if err == context.DeadlineExceeded {
					//可以超时重试
					if retryCnt > 10 {
						//中断业务执行
						return
					}
					ch <- struct{}{}
					continue
				}
				if err != nil {
					//不可挽回的错误 要考虑中断业务执行
					return
				}
				retryCnt = 0
			case <-stop:
				return
			}

		}
	}()

	stop <- struct{}{}
	//Output:
}

func (s *ClientE2ESuite) TestTryLock() {
	t := s.T()
	rdb := s.rdb
	client := &Client{
		client: rdb,
	}
	testCases := []struct {
		name string

		key        string
		expiration time.Duration

		wantLock *Lock
		wantErr  error

		before func()
		after  func()
	}{
		{
			// 加锁成功
			name:       "locked",
			key:        "locked-key",
			expiration: time.Minute,
			before:     func() {},
			after: func() {
				res, err := rdb.Del(context.Background(), "locked-key").Result()
				require.NoError(t, err)
				require.Equal(t, int64(1), res)
			},
			wantLock: &Lock{
				key: "locked-key",
			},
		},
		{
			// 模拟并发竞争失败
			name:       "failed",
			key:        "failed-key",
			expiration: time.Minute,
			before: func() {
				// 假设已经有人设置了分布式锁
				val, err := rdb.Set(context.Background(), "failed-key", "123", time.Minute).Result()
				require.NoError(t, err)
				require.Equal(t, "OK", val)
			},
			after: func() {
				res, err := rdb.Del(context.Background(), "failed-key").Result()
				require.NoError(t, err)
				require.Equal(t, int64(1), res)
			},
			wantErr: errs.NewErrFailedToPreemptLock(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			l, err := client.TryLock(context.Background(), tc.key, tc.expiration)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.key, l.key)
			assert.NotEmpty(t, l.value)
			tc.after()
		})
	}
}

func (s *ClientE2ESuite) TestUnLock() {
	t := s.T()
	rdb := s.rdb
	client := &Client{
		client: rdb,
	}
	testCases := []struct {
		name string

		lock *Lock

		before func()
		after  func()

		wantLock *Lock
		wantErr  error
	}{
		{
			name: "unlocked",
			lock: func() *Lock {
				l, err := client.TryLock(context.Background(), "unlocked-key", time.Minute)
				require.NoError(t, err)
				return l
			}(),
			before: func() {},
			after: func() {
				res, err := rdb.Exists(context.Background(), "unlocked-key").Result()
				require.NoError(t, err)
				require.Equal(t, int64(0), res)
			},
		},
		//{
		//	name: "lock not hold",
		//	lock: &Lock{
		//		key:    "not-hold-key",
		//		client: rdb,
		//		value:  "123",
		//	},
		//	wantErr: errs.NewErrLockNotHold(),
		//	before:  func() {},
		//	after:   func() {},
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			err := tc.lock.Unlock(context.Background())
			require.Equal(t, tc.wantErr, err)
			tc.after()
		})
	}
}
