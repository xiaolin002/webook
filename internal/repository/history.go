package repository

import (
	"context"
	"project/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, record domain.HistoryRecord) error
}

// 结构体和方法可自行补全 这里只是示例
