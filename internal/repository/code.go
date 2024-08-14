package repository

import (
	"context"
	"project/internal/repository/cache"
)

/**
 * @Description
 * @Date 2024/3/11 18:11
 **/

var ErrCodeVerifyTooMany = cache.ErrCodeVerifyToMany
var ErrCodeSendTooMany = cache.ErrCodeSendToMany

type CodeRepository interface {
	Send(ctx context.Context, biz string, phone, code string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}
type CacheCodeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(cac cache.CodeCache) CodeRepository {
	return &CacheCodeRepository{
		cache: cac,
	}

}

func (c *CacheCodeRepository) Send(ctx context.Context, biz string, phone, code string) error {
	return c.cache.Set(ctx, biz, phone, code)

}

func (c *CacheCodeRepository) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return c.cache.Verify(ctx, biz, phone, inputCode)

}
