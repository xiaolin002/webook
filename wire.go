//go:build wireinject

package main

import (
	"github.com/google/wire"
	"project/internal/events/article"
	"project/internal/repository"
	"project/internal/repository/cache"
	"project/internal/repository/dao"
	"project/internal/service"
	"project/internal/web"
	jwt2 "project/internal/web/jwt"
	"project/ioc"
)

/**
 * @Description
 * @Date 2024/3/12 19:02
 **/

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitSmsService,
		ioc.InitWechatService,
		ioc.InitConsumers,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,

		dao.NewUserDao,
		dao.NewArticleGORMDAO,
		dao.NewGORMInteractiveDAO,

		//kafka
		article.NewSaramaSyncProducer,
		article.NewInteractiveReadEventsConsumer,

		cache.NewUserCache,
		cache.NewCodeCache,
		cache.NewArticleRedisCache,
		cache.NewInteractiveRedisCache,

		repository.NewCacheUsersRepository,
		repository.NewCodeRepository,
		repository.NewCacheArticleRepository,
		repository.NewCachedInteractiveRepository,

		service.NewUsersService,
		service.NewCodeService,
		service.NewArticleService,
		service.NewInteractiveService,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		jwt2.NewRedisJWTHandler,

		ioc.InitWebServer,
		// gin中间件
		ioc.InitGinMiddlewares,

		wire.Struct(new(App), "*"),
	)
	return new(App)

}
