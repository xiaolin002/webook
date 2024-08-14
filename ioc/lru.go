package ioc

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	cac "project/internal/repository/cache"
	"time"
)

/**
 * @Description
 * @Date 2024/8/14 20:22
 **/

func InitLruCache() cac.CodeCache {
	cacheSize := 100 // 设置缓存大小
	cache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Error initializing cache:", err)
		return nil
	}

	expiration := time.Hour * 24 // 设置缓存过期时间
	return cac.NewLocalCodeCache(cache, expiration)

}
