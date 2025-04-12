package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"project/internal/domain"
	"project/internal/repository/cache"
	"project/internal/repository/dao"
	"time"
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
	// 仅自己可见
	SyncStatus(ctx context.Context, uid int64, id int64, private domain.ArticleStatus) error
	// GetByAuthor 作者自己查询自己的文章列表
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	// GetById 查看指定文章内容
	GetById(ctx context.Context, id int64) (domain.Article, error)
	// 读者查看指定的文章
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
}

type CacheArticleRepository struct {
	dao   dao.ArticleDAO
	cache cache.ArticleCache
	// userRepo 是为了获取作者的信息
	userRepo UserRepository
}

func NewCacheArticleRepository(dao dao.ArticleDAO, cache cache.ArticleCache, userRepo UserRepository) ArticleRepository {
	return &CacheArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
	}
}
func (c *CacheArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPubById(ctx, id)
	if err == nil {
		return res, nil
	}

	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 需要注意的是文章只有作者得id 没有作者的其他信息（name）
	// 所以需要去查询作者的信息
	res = c.ToDomain(dao.Article(art))
	author, err := c.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		// 可以返回res，err需要记录日志 ，因为吞掉了作者的名字请求错误
		return domain.Article{}, err
	}
	res.Author.Name = author.NickName
	// 回写缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = c.cache.SetPubById(ctx, res)
		if err != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (c *CacheArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.ToDomain(art)

	go func() {
		err = c.cache.Set(ctx, res)
		if err != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (c *CacheArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 策略是第一次的话就读redis 不是的话就读db，可根据limit和offset来判断
	// 如果作者的作品小于100那么也是可以设置的可以改为 limit <= 100
	if offset == 0 && limit == 100 {
		res, err2 := c.cache.GetFirstPage(ctx, uid)
		if err2 == nil {
			return res, nil
		} else {
			// 这里要记录日志
			//缓存未命中可以忽略
		}
	}

	arts, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return c.ToDomain(src)
	})

	// 同步 异步都行
	go func() {
		if offset == 0 && limit == 100 {
			// 如果走了db 就把数据写入redis
			// 错误两种 一是网络 二是redis本身的错误
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				// 这里要记录日志
				// 监控
			}
		}
	}()

	// 这里是制定的一个策略 当访问列表时 会访问第一个 所以设置第一个缓存
	// 异步的话最好设置一个新的context
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, res)
	}()

	return res, nil
}

func (c *CacheArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
	if err == nil {
		er := c.cache.DelFirstPage(ctx, uid)
		if er != nil {
			// 也要记录日志
		}
	}
	return err
}

func (c *CacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {

	id, err := c.dao.Insert(ctx, c.toEntity(art))
	if err == nil {
		// 由于作者新发布了文章，所以需要把缓存删除掉
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
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
	err := c.dao.UpdateById(ctx, c.toEntity(art))
	if err == nil {
		// 由于作者新发布了文章，所以需要把缓存删除掉
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	return err
}
func (c *CacheArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}

	// 规定新帖子一发布就回写缓存
	go func() {

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 你可以灵活设置过期时间
		user, er := c.userRepo.FindById(ctx, art.Author.Id)
		if er != nil {
			// 要记录日志
			return
		}
		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.NickName,
		}
		er = c.cache.SetPubById(ctx, art)
		if er != nil {
			// 记录日志
		}
	}()
	return id, err
}
func (c *CacheArticleRepository) ToDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			// 这里有一个错误
			Id: art.AuthorId,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

func (c *CacheArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	// 这里防止文章太大需要进行限制
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) <= size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			// 记录日志
		}

	}
}
