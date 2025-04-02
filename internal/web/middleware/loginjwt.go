package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	jwt2 "project/internal/web/jwt"
)

/**
 * @Description
 * @Date 2024/3/7 22:10
 **/

type LoginJwtMiddlewareBuilder struct {
	jwt2.Handler
}

func NewLoginJwtMiddlewareBuilder(hdl jwt2.Handler) *LoginJwtMiddlewareBuilder {
	return &LoginJwtMiddlewareBuilder{Handler: hdl}
}

func (m *LoginJwtMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/user/signup" || path == "/user/login" || path == "/user/login_sms/code/send" ||
			path == "/user/login_sms" || path == "/oauth2/wechat/authurl" || path == "/oauth2/wechat/callback" {
			return
		}
		// 被抽出来成了一个方法

		tokenStr := m.ExtractToken(ctx)

		var uc jwt2.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return jwt2.JWTKey, nil
		})
		if err != nil {
			// token 不对 伪造的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 不判定也行  到期时间 因为valid中已经做了校验
		if token == nil || !token.Valid {
			// token过期或者无效的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if uc.UserAgent != ctx.GetHeader("User-Agent") {

			//后期监控的时候这里要埋点  因为进这里大概率施工急着或者版本升级问题导致

			ctx.AbortWithStatus(http.StatusUnauthorized)
			return

		}
		err = m.CheckSession(ctx, uc.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("user", uc)

	}

}
