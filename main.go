package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"project/internal/web/middleware"
	"project/wire"
)

// 双击shift可查找任何东西
// ctrl +F 查找该文件中匹配的东西

func main() {
	initLogger()
	server := wire.InitWebServer()
	server.Run(":8081")

}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
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
