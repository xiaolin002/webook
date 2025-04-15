package ioc

import (
	"github.com/IBM/sarama"
	"project/internal/events"
	"project/internal/events/article"
)

// InitSaramaClient Kafka 客户端的初始化
func InitSaramaClient() sarama.Client {
	//type Config struct {
	//	Addr []string `yaml:"addr"`
	//}

	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"localhost:9094"}, scfg)
	if err != nil {
		panic(err)
	}
	return client
}

// InitSyncProducer 同步生产者的初始化
func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		panic(err)
	}
	return p
}

// InitConsumers 消费者列表的初始化    这里是一个切片，因为可能有多个消费者
func InitConsumers(c1 *article.InteractiveReadEventsConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
