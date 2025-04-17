package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"project/internal/service/sms"
	"time"
)

type Decorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

/*
  构造参数opt
	summaryOpts := prometheus.SummaryOpts{
		Name:       "sms_send_duration", // 指标名称
		Help:       "Duration of SMS send operations in milliseconds", // 帮助信息
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}, // 设定目标
	}
*/

func NewDecorator(svc sms.Service, opt prometheus.SummaryOpts) *Decorator {
	vector := prometheus.NewSummaryVec(opt, []string{"tpl_id"})
	prometheus.MustRegister(vector) // 注册到 Prometheus
	return &Decorator{
		svc:    svc,
		vector: vector,
	}
}

func (d *Decorator) Send(ctx context.Context,
	tplId string, args []string, numbers ...string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		d.vector.WithLabelValues(tplId).Observe(float64(duration))
	}()
	return d.svc.Send(ctx, tplId, args, numbers...)
}
