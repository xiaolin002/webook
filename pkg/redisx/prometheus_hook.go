package redisx

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"net"
	"strconv"
	"time"
)

/*
     可以对redis的命令进行监控
	summaryOpts := prometheus.SummaryOpts{
		Name:       "redis_command_duration", // 指标名称
		Help:       "Duration of Redis command operations in milliseconds", // 帮助信息
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}, // 设定目标
	}
	prometheusHook := NewPrometheusHook(summaryOpts)

	// 创建 Redis 客户端并注册钩子
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DialHooks: []redis.DialHook{
			prometheusHook.DialHook,
		},
		ProcessHooks: []redis.ProcessHook{
			prometheusHook.ProcessHook,
		},
		ProcessPipelineHooks: []redis.ProcessPipelineHook{
			prometheusHook.ProcessPipelineHook,
		},
	})

	// 使用 Redis 客户端
	ctx := context.Background()
	err := rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		log.Fatalf("Could not set key: %v", err)
	}

}

*/

type PrometheusHook struct {
	vector *prometheus.SummaryVec
}

func NewPrometheusHook(opt prometheus.SummaryOpts) *PrometheusHook {
	return &PrometheusHook{
		vector: prometheus.NewSummaryVec(opt, []string{"cmd", "key_exist"}),
	}
}

func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		var err error
		defer func() {
			duration := time.Since(start).Milliseconds()
			keyExists := err == redis.Nil
			p.vector.WithLabelValues(cmd.Name(), strconv.FormatBool(keyExists)).
				Observe(float64(duration))
		}()
		err = next(ctx, cmd)
		return err
	}
}

func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
