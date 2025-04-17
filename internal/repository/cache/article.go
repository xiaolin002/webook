package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"project/internal/domain"
	"time"
)

type ArticleCache interface {
	// 作者查询自己的文章列表接口页面的缓存
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, res []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error
	// 作者查看指定文章的接口的缓存
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, art domain.Article) error
	// 读者查看指定文章的接口的缓存  用pubKey为key
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
	SetPubById(ctx context.Context, res domain.Article) error
}

type ArticleRedisCache struct {
	client redis.Cmdable
}

func (a *ArticleRedisCache) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.pubKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (a *ArticleRedisCache) SetPubById(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.pubKey(art.Id), val, time.Minute*10).Err()
}

func NewArticleRedisCache(client redis.Cmdable) ArticleCache {
	return &ArticleRedisCache{
		client: client,
	}
}

func (a *ArticleRedisCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	key := a.firstKey(uid)
	//val, errs := a.client.Get(ctx, firstKey).Result()
	val, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}
func (a *ArticleRedisCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	// 这里你可以理解为设置缓存的时候只设置摘要 不设置内容
	// 所以篡改了一下，将内容设置为摘要
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	key := a.firstKey(uid)
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, key, val, time.Minute*10).Err()
}
func (a *ArticleRedisCache) DelFirstPage(ctx context.Context, uid int64) error {
	return a.client.Del(ctx, a.firstKey(uid)).Err()
}
func (a *ArticleRedisCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.key(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (a *ArticleRedisCache) Set(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.key(art.Id), val, time.Minute*10).Err()
}

func (a *ArticleRedisCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}

func (a *ArticleRedisCache) key(id int64) string {
	return fmt.Sprintf("article:detail:%d", id)
}

func (a *ArticleRedisCache) firstKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}

/*
编辑接口 ：

  修改
     ——————>删除缓存（delFirstPage）
  保存
发布接口：
  修改
     ——————>删除缓存 （delFirstPage）   同步到读者库    将文章设为热帖 加缓存SetPubById
  保存

仅自己可见：
  修改
     ——————>删除缓存（delFirstPage）   ---读者库和制作库全部更新

作者查看指定文章：

  设置缓存——————>Set()  Get()

作者查看自己的文章列表：
  设置缓存（将第一页设置为缓存）——————>SetFirstPage()  GetFirstPage()

读者查看指定文章：
  设置缓存——————>SetPubById()  GetPubById()



*/
