package article

import (
	"context"
	"github.com/IBM/sarama"
	"project/internal/repository"
	"project/pkg/saramax"
	"time"
)

type BatchInteractiveReadEventsConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
}

func NewBatchInteractiveReadEventsConsumer(repo repository.InteractiveRepository, client sarama.Client) *InteractiveReadEventsConsumer {
	return &InteractiveReadEventsConsumer{
		repo:   repo,
		client: client,
	}
}
func (i *BatchInteractiveReadEventsConsumer) StartV1() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			saramax.NewBatchHandler[ReadEvent](i.BatchConsume))
		if er != nil {
			// 记录日志
			//退出消费
		}
	}()
	return err
}
func (i *BatchInteractiveReadEventsConsumer) BatchConsume(msgs []*sarama.ConsumerMessage,
	events []ReadEvent) error {
	bizs := make([]string, 0, len(events))
	bizIds := make([]int64, 0, len(events))
	for _, evt := range events {
		bizs = append(bizs, "article")
		bizIds = append(bizIds, evt.Aid)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.BatchIncrReadCnt(ctx, bizs, bizIds)
}
