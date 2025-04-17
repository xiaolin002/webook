package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"project/internal/domain"
	"project/internal/service/oauth2/wechat"
	"time"
)

/**
 * @Description
 * @Date 2025/4/17 15:13
 **/

type Decorator struct {
	wechat.Service
	sum prometheus.Summary
}

/*
// NewSummary 创建并注册 Summary 指标
func NewSummary() prometheus.Summary {
	summaryOpts := prometheus.SummaryOpts{
		Name:       "wechat_verify_code_duration_summary",
		Help:       "Summary of the duration of WeChat verification code verification in milliseconds",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}
	summary := prometheus.NewSummary(summaryOpts)
	prometheus.MustRegister(summary)
	return summary
}


*/

func NewDecorator(svc wechat.Service, sum prometheus.Summary) *Decorator {
	return &Decorator{
		Service: svc,
		sum:     sum,
	}
}

func (d *Decorator) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		d.sum.Observe(float64(duration))
	}()
	return d.Service.VerifyCode(ctx, code)
}
