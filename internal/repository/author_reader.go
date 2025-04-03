package repository

import (
	"context"
	"project/internal/domain"
)

/**
 * @Description
 * @Date 2025/4/3 19:00
 **/

type ArticleReaderRepository interface {
	// Save 有则更新，无则插入，也就是 insert or update 语义
	Save(ctx context.Context, art domain.Article) error
}
