package saramax

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/xuhaidong1/go-generic-tools/pluginsx/logx"
	"time"
)

type BatchHandler[T any] struct {
	l           logx.Logger
	f           func(msg []*sarama.ConsumerMessage, t []T) error
	setupFunc   func(session sarama.ConsumerGroupSession) error
	cleanupFunc func(session sarama.ConsumerGroupSession) error
	batchSize   int
}

func (h *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	if h.setupFunc != nil {
		return h.setupFunc(session)
	}
	return nil
}

func (h *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	if h.cleanupFunc != nil {
		return h.cleanupFunc(session)
	}
	return nil
}

func (h *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgCh := claim.Messages()
	for {
		msgs := make([]*sarama.ConsumerMessage, 0, h.batchSize)
		ts := make([]T, 0, h.batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		done := false
		for i := 0; i < h.batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				//这一批超时，不需要尝试凑够一批了
				done = true
			case msg, ok := <-msgCh:
				if !ok {
					cancel()
					return nil
				}
				msgs = append(msgs, msg)
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					h.LogError("json反序列化失败", claim, err)
					session.MarkMessage(msg, "")
					continue
				}
				ts = append(ts, t)
			}
		}
		err := h.f(msgs, ts)
		if err != nil {
			//在业务逻辑里面处理错误,err!=nil就不提交
			h.LogError("业务消息处理错误", claim, err)
			continue
		}
		for _, msg := range msgs {
			session.MarkMessage(msg, "")
		}
		cancel()
	}
}

func NewBatchHandler[T any](l logx.Logger, batchSize int, f func(msg []*sarama.ConsumerMessage, t []T) error, opts ...Option[T]) *BatchHandler[T] {
	b := &BatchHandler[T]{l: l, f: f, batchSize: batchSize}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (h *BatchHandler[T]) LogError(logMsg string, claim sarama.ConsumerGroupClaim, err error) {
	h.l.Error(logMsg,
		logx.String("topic", claim.Topic()),
		logx.Int32("partition", claim.Partition()),
		logx.Error(err))
}

type Option[T any] func(h *BatchHandler[T])

func WithSetup[T any](f func(session sarama.ConsumerGroupSession) error) Option[T] {
	return func(h *BatchHandler[T]) {
		h.setupFunc = f
	}
}

func WithCleanup[T any](f func(session sarama.ConsumerGroupSession) error) Option[T] {
	return func(h *BatchHandler[T]) {
		h.cleanupFunc = f
	}
}
