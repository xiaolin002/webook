package repository

import (
	"context"
	"github.com/gin-gonic/gin"
	"project/internal/domain"
	"project/internal/repository/dao"
)

/**
 * @Description
 * @Date 2025/4/3 17:03
 **/

type ArticleRepository interface {
	// create 和update都是作者在自己的地方编辑然后存储 未发布
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// Sync 同步文章，有则更新，无则插入，也就是 insert or update 语义
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx *gin.Context, uid int64, id int64, private domain.ArticleStatus) error
}

type CacheArticleRepository struct {
	dao dao.ArticleDAO
}

func (c *CacheArticleRepository) SyncStatus(ctx *gin.Context, uid int64, id int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
	//if err == nil {
	//	er := c.cache.DelFirstPage(ctx, uid)
	//	if er != nil {
	//		// 也要记录日志
	//	}
	//}
	return err
}

func NewCacheArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &CacheArticleRepository{dao: dao}
}

func (c *CacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(art))
	return id, err
}
func (c *CacheArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}
func (c *CacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
	// 根据文章的id来更新
	return c.dao.UpdateById(ctx, c.toEntity(art))
}
func (c *CacheArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	return id, err
}
