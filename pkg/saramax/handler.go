package saramax

import (
	"encoding/json"
	"github.com/IBM/sarama"
)

type Handler[T any] struct {
	fn func(msg *sarama.ConsumerMessage, event T) error
}

func NewHandler[T any](fn func(msg *sarama.ConsumerMessage, event T) error) *Handler[T] {
	return &Handler[T]{
		fn: fn,
	}
}

// 实现 ConsumerGroupHandler

func (h *Handler[T]) Setup(sess sarama.ConsumerGroupSession) error {
	//在消费者组开始消费消息之前调用，你可以在这个方法里进行资源初始化、日志记录等操作。
	return nil
}

func (h *Handler[T]) Cleanup(sess sarama.ConsumerGroupSession) error {
	//在消费者组停止消费消息之后调用，用于执行一些清理操作。释放资源、保存状态
	return nil
}

func (h *Handler[T]) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	//这是处理消息的核心方法，用于消费分配给当前消费者的分区消息
	msgs := claim.Messages()
	for msg := range msgs {
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			// 记录日志
			//引入重试
			continue
		}
		err = h.fn(msg, t)
		if err != nil {
			// 记录日志
		}

		//处理完成后，需要调用 sess.MarkMessage 来标记消息已被处理
		sess.MarkMessage(msg, "")
	}
	return nil

}
