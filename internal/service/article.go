package service

import (
	"context"
	"project/internal/domain"
	"project/internal/repository"
)

/**
 * @Description
 * @Date 2025/4/3 16:12
 **/

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
}
type articleService struct {
	repo repository.ArticleRepository
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

// 编辑和创建在这里就需要进行区分了

// Save 编辑或者创建
func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	if art.Id > 0 {
		err := a.update(ctx, art)
		return art.Id, err
	}
	return a.create(ctx, art)
}

func (a *articleService) update(ctx context.Context, art domain.Article) error {
	return a.repo.Update(ctx, art)
}

func (a *articleService) create(ctx context.Context, art domain.Article) (int64, error) {
	return a.repo.Create(ctx, art)
}
func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	return a.repo.Sync(ctx, art)

}
