package article

import (
	"encoding/json"
	"github.com/IBM/sarama"
)

const TopicReadEvent = "article_read"

type Producer interface {
	ProduceReadEvents(evt ReadEvent) error
}

type ReadEvent struct {
	// 文章ID
	Aid int64
	// 用户ID
	Uid int64
}

type SaramaSyncProducer struct {
	// 用的是同步
	producer sarama.SyncProducer
}

func NewSaramaSyncProducer(producer sarama.SyncProducer) Producer {
	return &SaramaSyncProducer{
		producer: producer,
	}
}

func (s *SaramaSyncProducer) ProduceReadEvents(evt ReadEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicReadEvent,
		Value: sarama.StringEncoder(val),
	})
	return err

}
