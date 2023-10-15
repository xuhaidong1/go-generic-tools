package queue

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

// 实现了抢到锁开始生产，锁丢掉停止生产，接受http请求开始或停止生产，服务关闭停止生产及关闭所有监听和工作goroutine的操作
func TestRadioChan_Broadcast(t *testing.T) {
	startCond, stopCond := NewCondAtomic(&sync.Mutex{}), NewCondAtomic(&sync.Mutex{})
	ctx, cancel := context.WithCancel(context.Background())
	_ = NewProducer(ctx, startCond, stopCond)
	_ = NewLoadBalancer(ctx, startCond, stopCond)
	startCond.L.Lock()
	startCond.Broadcast()
	startCond.L.Unlock()
	time.Sleep(time.Second)
	stopCond.L.Lock()
	stopCond.Broadcast()
	stopCond.L.Unlock()
	time.Sleep(time.Second)
	startCond.L.Lock()
	startCond.Broadcast()
	startCond.L.Unlock()
	time.Sleep(time.Second)
	cancel()
	time.Sleep(time.Second)
}

type Producer struct {
	startCond *CondAtomic
	stopCond  *CondAtomic
	worker    *ProduceWorker
}

type ProduceWorker struct {
	cancel context.CancelFunc
}

func NewProduceWorker(ctx context.Context) *ProduceWorker {
	workerCtx, cancel := context.WithCancel(ctx)
	p := &ProduceWorker{
		cancel: cancel,
	}
	go p.Work(workerCtx)
	return p
}

func (w *ProduceWorker) Work(ctx context.Context) {
	once := sync.Once{}
	for {
		select {
		case <-ctx.Done():
			log.Println("producer work结束")
			return
		default:
			once.Do(func() {
				log.Println("producer working")
			})
		}
	}
}

func (w *ProduceWorker) StopWork() {
	w.cancel()
}

func NewProducer(ctx context.Context, start, stop *CondAtomic) *Producer {
	p := &Producer{startCond: start, stopCond: stop}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go p.listenStartCond(ctx, wg)
	go p.listenStopCond(ctx, wg)
	wg.Wait()
	return p
}

func (p *Producer) listenStartCond(ctx context.Context, wg *sync.WaitGroup) {
	wg.Done()
	for {
		p.startCond.L.Lock()
		err := p.startCond.WaitWithTimeout(ctx)
		p.startCond.L.Unlock()
		if err != nil {
			log.Println("producer listenStartCond closing")
			return
		}
		if p.worker == nil {
			p.worker = NewProduceWorker(ctx)
		}
	}
}

// 这边收到了停止信号，是不知道什么原因让停止的 /没拿到锁 应该传黑匣子/手动停止 应该传黑匣子/服务关闭 应该传黑匣子--统一了 不需要知道原因
func (p *Producer) listenStopCond(ctx context.Context, wg *sync.WaitGroup) {
	wg.Done()
	for {
		p.stopCond.L.Lock()
		err := p.stopCond.WaitWithTimeout(ctx)
		p.stopCond.L.Unlock()
		if err != nil {
			log.Println("producer listenStopCond closing")
			return
		}
		if p.worker != nil {
			p.worker.StopWork()
			p.worker = nil
		}
	}
}

type LoadBalancer struct {
	startCond *CondAtomic
	stopCond  *CondAtomic
	worker    *LoadBalanceWorker
}

type LoadBalanceWorker struct {
	cancel context.CancelFunc
}

func NewLoadBalanceWorker(ctx context.Context) *LoadBalanceWorker {
	workerCtx, cancel := context.WithCancel(ctx)
	w := &LoadBalanceWorker{
		cancel: cancel,
	}
	go w.Work(workerCtx)
	return w
}

func (w *LoadBalanceWorker) Work(ctx context.Context) {
	once := sync.Once{}
	for {
		select {
		case <-ctx.Done():
			log.Println("loadbalancer work结束")
			return
		default:
			once.Do(func() {
				log.Println("loadbalancer working")
			})
		}
	}
}

func (w *LoadBalanceWorker) StopWork() {
	w.cancel()
}

func NewLoadBalancer(ctx context.Context, start, stop *CondAtomic) *LoadBalancer {
	l := &LoadBalancer{startCond: start, stopCond: stop}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go l.listenStartCond(ctx, wg)
	go l.listenStopCond(ctx, wg)
	wg.Wait()
	return l
}

func (l *LoadBalancer) listenStartCond(ctx context.Context, wg *sync.WaitGroup) {
	wg.Done()
	for {
		l.startCond.L.Lock()
		err := l.startCond.WaitWithTimeout(ctx)
		l.startCond.L.Unlock()
		if err != nil {
			log.Println("loadbalancer listenStartCond closing")
			return
		}
		if l.worker == nil {
			l.worker = NewLoadBalanceWorker(ctx)
		}
	}
}

func (l *LoadBalancer) listenStopCond(ctx context.Context, wg *sync.WaitGroup) {
	wg.Done()
	for {
		l.stopCond.L.Lock()
		err := l.stopCond.WaitWithTimeout(ctx)
		l.stopCond.L.Unlock()
		if err != nil {
			log.Println("producer listenStopCond closing")
			return
		}
		if l.worker != nil {
			l.worker.StopWork()
			l.worker = nil
		}
	}
}
