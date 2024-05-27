package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"project/internal/repository"
	"project/internal/repository/dao"
	"project/internal/service"
	"project/internal/web"
	"project/internal/web/middleware"
	"project/pkg/ginx/middliware/ratelimit"
	"strings"
	"time"
)

// 双击shift可查找任何东西
// ctrl +F 查找该文件中匹配的东西

func main() {

	db := initDB()
	router := initWebServer()
	initUserHdl(db, router)
	err := router.Run(":8081")
	if err != nil {
		panic("端口可能被占用")
	}
	// 使用这种写法
	//hdl := &user2.UserHandler{}
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDao(db)
	repo := repository.NewUsersRepository(ud)
	svc := service.NewUsersService(repo)
	hdl := web.NewUserHandler(svc)
	hdl.RegisterRouter(server)
}

func initDB() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/xiaohongshu?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库驱动错误")
	}
	if err := dao.InitTables(db); err != nil {
		panic("创建数据库表错误")
	}
	return db
}
func initWebServer() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"}, // 这里的Authorization代表的是jwt中的头部
		ExposeHeaders:    []string{"x-jwt-token"},                   // 允许前端能够访问后端响应中带的头部 ，一般在公司中可能加一些自定义头部，依次在这里加就行
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}))
	useJwtSession(router)
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	// 会限流 报429错误http状态码
	router.Use(ratelimit.NewBuilder(redisClient, time.Minute, 100).Build())

	return router
}
func useSession(server *gin.Engine) {
	//创建基于cookie的存储引擎，参数是用于加密的密钥
	store := cookie.NewStore([]byte("secret"))
	// 基于内存的实现  32位或者64位的随机密钥，最好不要用特殊符号,有时可能识别不了
	//store := memstore.NewStore([]byte(""), []byte(""))
	// 基于redis实现  // 最后byte 身份认证和数据加密 这两者加上授权就是信息安全的三个核心概念
	//store, err := redis.NewStore(5, "tcp", "39.105.211.136:6379", "",
	//	[]byte("rX6`tC9[hP5:nY0#eW3_lK3]eV5@zO3>"), []byte("jI2.hR2:vC6~uV3;cQ1_wV3:mK5$nL5."))
	//if err != nil {
	//	panic(err)
	//}

	server.Use(sessions.Sessions("ssid", store))
	login := &middleware.LoginMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}
func useJwtSession(server *gin.Engine) {
	login := middleware.LoginJwtMiddlewareBuilder{}
	server.Use(login.CheckLogin())

	/*
	   优缺点：
	   1.不依赖三方存储（提高性能）
	   2. 适合分布式

	   1. 对加密依赖大容易泄密
	    2. 最好不要在jwt里放敏感信息
	*/

}
