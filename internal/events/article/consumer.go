package article

import (
	"context"
	"github.com/IBM/sarama"
	"project/internal/repository"
	"project/pkg/saramax"
	"time"
)

/**
 * @Description
 * @Date 2025/4/15 20:00
 **/

type InteractiveReadEventsConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
}

func NewInteractiveReadEventsConsumer(repo repository.InteractiveRepository, client sarama.Client) *InteractiveReadEventsConsumer {
	return &InteractiveReadEventsConsumer{
		repo:   repo,
		client: client,
	}
}
func (i *InteractiveReadEventsConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{TopicReadEvent}, saramax.NewHandler(i.Consume))
		if er != nil {
			// 记录日志
			//退出消费
		}

	}()

	return err
}

func (i *InteractiveReadEventsConsumer) Consume(msg *sarama.ConsumerMessage, event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", event.Aid)

}
