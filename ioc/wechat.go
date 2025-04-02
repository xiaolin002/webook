package ioc

import (
	"os"
	"project/internal/service/oauth2/wechat"
)

/**
 * @Description
 * @Date 2025/4/1 21:11
 **/

func InitWechatService() wechat.Service {
	appID, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("找不到环境变量 WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("找不到环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewAuthService(appID, appSecret)
}
