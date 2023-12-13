package saramax

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/xuhaidong1/go-generic-tools/pluginsx/logx"
)

type Handler[T any] struct {
	l logx.Logger
	f func(msg *sarama.ConsumerMessage, t T) error
}

func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			h.LogError("json反序列化消息失败", msg, err)
			//不中断，直接提交错误的消息
			session.MarkMessage(msg, "")
			continue
		}
		err = h.f(msg, t)
		if err != nil {
			h.LogError("处理消息失败", msg, err)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func NewHandler[T any](l logx.Logger, f func(msg *sarama.ConsumerMessage, t T) error) *Handler[T] {
	return &Handler[T]{l: l, f: f}
}

func (h *Handler[T]) LogError(logMsg string, msg *sarama.ConsumerMessage, err error) {
	h.l.Error(logMsg,
		logx.String("topic", msg.Topic),
		logx.Int32("partition", msg.Partition),
		logx.Int64("offset", msg.Offset),
		logx.Error(err))
}
