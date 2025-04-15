package article

import (
	"context"
	"github.com/IBM/sarama"
	"project/internal/domain"
	"project/internal/repository"
	"project/pkg/saramax"
	"time"
)

// 阅读记录事件 可补全

type HistoryRecordConsumer struct {
	repo   repository.HistoryRecordRepository
	client sarama.Client
}

func (i *HistoryRecordConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			// 这里的没有生产者事件 正常需要替换
			saramax.NewHandler[ReadEvent](i.Consume))
		if er != nil {
			// 记录日志
			//退出消费
		}
	}()
	return err
}

func (i *HistoryRecordConsumer) Consume(msg *sarama.ConsumerMessage,
	event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.AddRecord(ctx, domain.HistoryRecord{
		BizId: event.Aid,
		Biz:   "article",
		Uid:   event.Uid,
	})
}
