package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

/**
 * @Description
 * @Date 2024/3/6 22:37
 **/

type LoginMiddlewareBuilder struct {
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	// 注册一下这个类型
	gob.Register(time.Now())

	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/user/signup" || path == "/user/login" || path == "/user/login_sms/code/send" || path == "/user/login_sms" {
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("uid")
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 保持登录态
		now := time.Now()
		const updateTimeKey = "update_time"
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if val == nil || !ok || now.Sub(lastUpdateTime) > time.Minute {
			sess.Set(updateTimeKey, now) // 注意这里的now是时间戳 而redis是字节切片 故需要注册一下
			sess.Set("uuid", userId)
			err := sess.Save()
			if err != nil {
				//打印日志
				log.Println(err)
			}

		}

	}
}
