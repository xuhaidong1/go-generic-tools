package redis_lock

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/xuhaidong1/go-generic-tools/container/queue"
	"github.com/xuhaidong1/go-generic-tools/redis_lock/errs"
	"log"
	"testing"
	"time"
)

type ClientE2ESuite struct {
	suite.Suite
	rdb redis.Cmdable
}

func (s *ClientE2ESuite) SetupSuite() {
	s.rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6579",
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

func TestProducerRefresh(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6579",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	lockClient := NewClient(rdb)
	l, err := lockClient.TryLock(context.Background(), "test_producer_refresh", "123456", time.Second*2)
	if err != nil {
		return
	}
	//超时重试控制信号
	ch := make(chan struct{}, 1)
	stop := make(chan struct{})
	bizStop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)
		log.Println("a")
		//不断续约 直到收到退出信号
		defer ticker.Stop()
		//重试次数计数
		retryCnt := 0
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
				err := l.Refresh(ctx)
				cancel()
				if err == context.DeadlineExceeded {
					//可以超时重试
					log.Println("超时")
					ch <- struct{}{}
					continue
				}
				if err != nil {
					//不可挽回的错误 要考虑中断业务执行
					bizStop <- struct{}{}
					return
				}
				retryCnt = 0
			case <-ch:
				retryCnt++
				ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
				err := l.Refresh(ctx)
				cancel()
				if err == context.DeadlineExceeded {
					//可以超时重试
					if retryCnt > 3 {
						//中断业务执行
						bizStop <- struct{}{}
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
				er := l.Unlock(context.Background())
				log.Println(er)
				return
			}
		}
	}()
	//Output:
	time.Sleep(time.Second * 5)
	stop <- struct{}{}
}

// 模拟消费者抢生产者的锁
//func TestTryProducerLock(t *testing.T) {
//	wg := sync.WaitGroup{}
//	wg.Add(3)
//
//	rdb := redis.NewClient(&redis.Options{
//		Addr:     "localhost:6579",
//		Password: "", // no password set
//		DB:       0,  // use default DB
//	})
//	//rdb.Set(context.Background(), "test_producer_refresh", "1234", time.Second*2)
//	//lockClient := NewClient(rdb)
//	//l, err := lockClient.TryLock(context.Background(), "test_producer_refresh", "1234", time.Second*2)
//	//if err != nil {
//	//	return
//	//}
//	//log.Println("1成为了生产者")
//	//stopProduceCh := make(chan struct{})
//	//producerCtx, producerCtxCancel := context.WithCancel(context.Background())
//	//go func() {
//	//	er := ProducerRefreshLock(producerCtx, l)
//	//	if er == errors.New("生产者续约失败，需要中断生产") {
//	//		log.Println("生产者续约失败，需要中断生产")
//	//		stopProduceCh <- struct{}{}
//	//		//另一边有worker监听这个 stop就不生产了
//	//		wg.Done()
//	//		return
//	//	}
//	//	log.Println("1卸任了生产者...")
//	//	wg.Done()
//	//	return
//	//}()
//	//go func() {
//	//	ticker := time.NewTicker(time.Second)
//	//	//不断续约 直到收到退出信号
//	//	defer ticker.Stop()
//	//	//重试次数计数
//	//	retryCnt := 0
//	//	for {
//	//		select {
//	//		case <-ticker.C:
//	//			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
//	//			err := l.Refresh(ctx)
//	//			cancel()
//	//			if err == context.DeadlineExceeded {
//	//				//可以超时重试
//	//				log.Println("超时")
//	//				ch <- struct{}{}
//	//				continue
//	//			}
//	//			if err != nil {
//	//				//不可挽回的错误 要考虑中断业务执行
//	//				bizStop <- struct{}{}
//	//				wg.Done()
//	//				return
//	//			}
//	//			retryCnt = 0
//	//		case <-ch:
//	//			retryCnt++
//	//			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
//	//			err := l.Refresh(ctx)
//	//			cancel()
//	//			if err == context.DeadlineExceeded {
//	//				//可以超时重试
//	//				if retryCnt > 3 {
//	//					//中断业务执行
//	//					bizStop <- struct{}{}
//	//					return
//	//				}
//	//				ch <- struct{}{}
//	//				continue
//	//			}
//	//			if err != nil {
//	//				//不可挽回的错误 要考虑中断业务执行
//	//				wg.Done()
//	//				return
//	//			}
//	//			retryCnt = 0
//	//		case <-stop:
//	//			_ = l.Unlock(context.Background())
//	//			log.Println("1卸任了生产者")
//	//			wg.Done()
//	//			return
//	//		}
//	//	}
//	//}()
//	//-------以上是生产者在续约锁
//	//----下面是消费者抢生产者的锁，生产者掉了，本消费者成为生产者
//	go func() {
//		lockKey := "test_producer_refresh"
//		podName := "offlinePush-2"
//		expiration := time.Second * 3
//		rdb := redis.NewClient(&redis.Options{
//			Addr:     "localhost:6579",
//			Password: "", // no password set
//			DB:       0,  // use default DB
//		})
//		lockClient2 := NewClient(rdb)
//		ProducerCtx, ProducerCtxCancel := context.WithCancel(context.Background())
//		startProduceCh := make(chan struct{})
//		stopProduceCh := make(chan struct{})
//		go func() {
//			lock, err := ApplyProducer(ProducerCtx, lockClient2, lockKey, podName, expiration)
//			if err != nil {
//				log.Fatal(err)
//				return
//			}
//			startProduceCh <- struct{}{}
//			err = ProducerRefreshLock(ProducerCtx, lock, time.Millisecond*300)
//			if errors.Is(err, errors.New("生产者续约失败，需要中断生产")) {
//				log.Println("生产者续约失败，需要中断生产")
//				stopProduceCh <- struct{}{}
//				//另一边有worker监听这个 stop就不生产了
//				wg.Done()
//				return
//			}
//			log.Println("1卸任了生产者...")
//			wg.Done()
//			return
//		}()
//		select {
//		case <-startProduceCh:
//			log.Printf("%v开始生产", podName)
//		case <-stopProduceCh:
//			log.Printf("%v停止生产", podName)
//		case <-ProducerCtx.Done():
//			log.Printf("%v停止生产", podName)
//		}
//		ProducerCtxCancel()
//	}()
//	go func() {
//
//	}()
//	time.Sleep(time.Second * 5)
//	producerCtxCancel()
//	wg.Wait()
//}

// 描述抢分布式锁的超时策略
type ApplyLockRetry struct {
	retryCnt int
	maxCnt   int
}

func NewApplyLockRetry() RetryStrategy {
	return &ApplyLockRetry{
		retryCnt: 0,
		maxCnt:   3,
	}
}
func (r *ApplyLockRetry) Next() (time.Duration, bool) {
	r.retryCnt++
	if r.retryCnt >= r.maxCnt {
		r.retryCnt = 0
		return time.Millisecond * 300, false
	}
	return time.Millisecond * 300, true
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
		val        string
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
			val:        "locked-val",
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
			val:        "failed-val",
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
			wantErr: errs.ErrFailedToPreemptLock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			l, err := client.TryLock(context.Background(), tc.key, tc.val, tc.expiration)
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
				l, err := client.TryLock(context.Background(), "unlocked-key", "unlocked-val", time.Minute)
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

//func TestProducer(t *testing.T) {
//	lockKey := "test_producer_refresh"
//	podName := "offlinePush-2"
//	expiration := time.Second * 3
//	rdb := redis.NewClient(&redis.Options{
//		Addr:     "localhost:6579",
//		Password: "", // no password set
//		DB:       0,  // use default DB
//	})
//	lockClient2 := NewClient(rdb)
//	ProducerCtx, ProducerCtxCancel := context.WithCancel(context.Background())
//	startProduceCh := make(chan struct{})
//	stopProduceCh := make(chan struct{})
//	go func() {
//		lock, err := ApplyProducer(ProducerCtx, lockClient2, lockKey, podName, expiration)
//		if err != nil {
//			log.Fatal(err)
//			return
//		}
//		startProduceCh <- struct{}{}
//		err = ProducerRefreshLock(ProducerCtx, lock, time.Millisecond*300)
//		if errors.Is(err, errors.New("生产者续约失败，需要中断生产")) {
//			log.Println("生产者续约失败，需要中断生产")
//			stopProduceCh <- struct{}{}
//			//另一边有worker监听这个 stop就不生产了
//			wg.Done()
//			return
//		}
//		log.Println("1卸任了生产者...")
//		wg.Done()
//		return
//	}()
//	select {
//	case <-startProduceCh:
//		log.Printf("%v开始生产", podName)
//	case <-stopProduceCh:
//		log.Printf("%v停止生产", podName)
//	case <-ProducerCtx.Done():
//		log.Printf("%v停止生产", podName)
//	}
//	ProducerCtxCancel()
//}

type LockManager struct {
	lockClient *Client
	//要抢的锁的名字
	lockKey string
	//服务实例名
	podName string
	//持有锁的过期时间
	expiration time.Duration
	//续约的时间间隔
	refreshInterval time.Duration
	//redis lua脚本执行超时时间
	timeout time.Duration
	//申请到的锁
	l *Lock
	//根据有没有锁/推送开关开没开 控制业务（produceworker，loadbalanceworker）的启动停止
	startCond *queue.CondAtomic
	stopCond  *queue.CondAtomic
	cancel    context.CancelFunc
}

// 需要在外面开启goroutine
func (m *LockManager) AutoRefresh(ctx context.Context, l *Lock, timeout time.Duration) error {
	ticker := time.NewTicker(time.Second)
	//不断续约 直到收到退出信号
	defer ticker.Stop()
	//重试次数计数
	retryCnt := 0
	//超时重试控制信号
	retryCh := make(chan struct{}, 1)
	for {
		select {
		case <-ticker.C:
			refreshCtx, cancel := context.WithTimeout(ctx, timeout)
			err := l.Refresh(refreshCtx)
			cancel()
			if err == context.DeadlineExceeded {
				//可以超时重试
				log.Println("超时")
				retryCh <- struct{}{}
				continue
			}
			if err != nil {
				//不可挽回的错误 要考虑中断业务执行，应该先中断生产，后解锁，
				//如果先解锁，其它实例成为生产者，我们还没中断生产，生产进度还没写回到redis，新生产者就开始生产了，会导致生产两份相同的消息
				//应该先中断生产，生产黑匣子写回redis，解锁
				//生产写回redis&派任务时收到停止生产信号，应该无视，下一轮发现ctx没锁了，就退出了
				//新生产者上任：检查上一个生产者的黑匣子（一个指定key：producer_black_box）（记录了意外中断时在生产但还没有写回redis的business名字）（应该只有1个），生产日期time），
				//把这些黑匣子里的内容都重新生产一遍。
				//生产者正常写回缓存时应该一股脑写到redis，但负载均衡应该按照本地缓存里面的分片的消息一一分配，完成一个business的生产-写redis-对本地缓存做负载均衡-一片数据负载均衡一次-完成分配就清除本地缓存
				return errors.New("生产者续约失败，需要中断生产")
			}
			retryCnt = 0
		case <-retryCh:
			retryCnt++
			refreshCtx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(refreshCtx)
			cancel()
			if err == context.DeadlineExceeded {
				//可以超时重试
				if retryCnt > 3 {
					return errors.New("生产者续约失败，需要中断生产")
				}
				//如果retryCh容量不是1，在这会阻塞
				retryCh <- struct{}{}
				continue
			}
			if err != nil {
				//不可挽回的错误 要考虑中断业务执行
				return errors.New("生产者续约失败，需要中断生产")
			}
			retryCnt = 0
		case <-ctx.Done():
			//生产者主动退出
			if ctx.Err() == context.Canceled {
				//log.Println("卸任了生产者")
				return nil
			}
			return nil
		}
	}
}

// 外面拿到lock可以建立stopProduceCh 或者ctx
// ProducerCtx, ProducerCtxCancel := context.WithCancel(context.Background())
// 注意client要传重试策略
func (m *LockManager) Apply(ctx context.Context) (*Lock, error) {
	//ctx 申请producer用的ctx，可以级联取消
	//每隔一秒申请一次 直到收到退出信号
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l, err := m.lockClient.TryLock(ctx, m.lockKey, m.podName, m.expiration)
			if err != nil && !errors.Is(err, errs.ErrLockNotHold) {
				log.Printf("%v抢生产者失败\n", m.podName)
				return nil, err
			}
			if errors.Is(err, errs.ErrLockNotHold) {
				log.Printf("%v没抢到生产者\n", m.podName)
				continue
			}
			log.Printf("%v成为了生产者", m.podName)
			return l, nil
		case <-ctx.Done():
			//生产者主动退出
			return nil, context.Canceled
		}
	}
}

// goroutine调用这个函数
func (m *LockManager) StartWork(ctx context.Context) {
	lock, err := m.Apply(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
		return
	}
	if errors.Is(err, context.Canceled) {
		return
	}
	m.l = lock
	m.startCond.L.Lock()
	m.startCond.Broadcast()
	m.startCond.L.Unlock()
	err = m.AutoRefresh(ctx, lock, time.Millisecond*300)
	//todo 错误处理细化 丢锁，手动停止，服务关闭
	if errors.Is(err, errors.New("生产者续约失败，需要中断生产")) {
		log.Println("生产者续约失败，需要中断生产")
		m.stopCond.L.Lock()
		m.stopCond.Broadcast()
		m.stopCond.L.Unlock()
	}
	_ = m.l.Unlock(ctx)
	m.l = nil
	return
}

func (m *LockManager) StopWork() {
	if m.l != nil {
		_ = m.l.Unlock(context.Background())
	}
	m.cancel()
}
