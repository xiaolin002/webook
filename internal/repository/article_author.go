package repository

import (
	"context"
	"project/internal/domain"
)

/**
 * @Description
 * @Date 2025/4/3 19:00
 **/

type ArticleAuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
}
