package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"project/internal/web"
	jwt2 "project/internal/web/jwt"
	"project/internal/web/middleware"
	"project/pkg/ginx"
	"project/pkg/ginx/middliware/prometheus"
	"project/pkg/ginx/middliware/ratelimit"
	"strings"
	"time"
)

/**
 * @Description  这里主要是对中间件和路由进行统一封装然后利用wire进行管理
 * @Date 2023/11/25 20:59
 **/

func InitWebServer(mdls []gin.HandlerFunc, hdl *web.UsersHandler,
	wechatHdl *web.OAuth2WechatHandler, artHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoute(server)
	wechatHdl.RegisterRoute(server)
	artHdl.RegisterRoute(server)
	return server
}

// InitWebServerV1 这里将UserHandler 抽出了接口 两种写法都对 语言方面可能不一样java可能这样写

//func InitWebServer(mdls []gin.HandlerFunc, hdls []web.Handler) *gin.Engine {
//	server := gin.Default()
//	server.Use(mdls...)
//	for _, hdl := range hdls {
//		hdl.RegisterRoutes(server)
//	}
//
//	return server
//}

func InitGinMiddlewares(redisClient redis.Cmdable, hdl jwt2.Handler) []gin.HandlerFunc {
	pb := &prometheus.Builder{
		Namespace: "geektime_daming",
		Subsystem: "webook",
		Name:      "gin_http",
		Help:      "统计 GIN 的HTTP接口数据",
	}
	ginx.InitCounter(prometheus2.CounterOpts{
		Namespace: "geektime_daming",
		Subsystem: "webook",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	return []gin.HandlerFunc{
		// 跨域请求中间件
		cors.New(cors.Config{
			AllowCredentials: true,
			AllowHeaders:     []string{"Content-Type", "Authorization"},  // 这里的Authorization代表的是jwt中的头部
			ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"}, // 允许前端能够访问后端响应中带的头部 ，一般在公司中可能加一些自定义头部，依次在这里加就行
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "http://localhost") {
					return true
				}
				return strings.Contains(origin, "your_company.com")
			},
			MaxAge: 12 * time.Hour,
		}),
		// jwt中间件
		//(&middleware.LoginJwtMiddlewareBuilder{}).CheckLogin(),
		pb.BuildResponseTime(),
		pb.BuildActiveRequest(),
		// 限流中间件
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
		// 使用session中间件
		//sessionHandlerFunc(), // 这里应该是初始化session store用的 否则无法用session
		middleware.NewLoginJwtMiddlewareBuilder(hdl).CheckLogin(),
		//middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al middleware.AccessLog) {
		// 自定义日志打印
		//	l.Debug("", logger.Field{Key: "req", Val: al})
		//logEntry := fmt.Sprintf("Time: %s, Path: %s, Method: %s, Status: %d, Duration: %s, ReqBody: %s, RespBody: %s\n",
		//			time.Now().Format(time.RFC3339), l.Path, l.Method, l.Status, l.Duration, l.ReqBody, l.RespBody)
		//  l.Debug("", logger.Field{Key: "req log", Val: logEntry})
		//}).AllowReqBody().AllowRespBody().Build(),
	}

}
