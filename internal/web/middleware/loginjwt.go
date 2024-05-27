package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"project/internal/web"
	"strings"
	"time"
)

/**
 * @Description
 * @Date 2024/3/7 22:10
 **/

type LoginJwtMiddlewareBuilder struct {
}

func (m *LoginJwtMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/user/signup" || path == "/user/login" {
			return
		}
		/*
		   1. 获取jwt的token
		   2. 验证token
		   3. 如果过期重新弄一个token
		*/
		//  1. 获取jwt的token
		tokenCode := ctx.GetHeader("Authorization")
		if tokenCode == "" {
			// 没有数据没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 因为 前端待会的数据是以 Bear  ****形式的，故需要分割
		seg := strings.Split(tokenCode, " ")
		if len(seg) > 2 {
			// 非法的token
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := seg[1]
		// 验证token
		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.Jwtkey, nil
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

		//时间过期需要重新设置
		expireTime := uc.ExpiresAt
		if expireTime.Sub(time.Now()) < time.Second*50 {
			expireTime = jwt.NewNumericDate(time.Now().Add(time.Second))
			tokenStr, err = token.SignedString(web.Jwtkey)
			ctx.Header("x-jwt-token", tokenStr)
			if err != nil {
				log.Println(err)
			}

		}
		ctx.Set("user", uc)

	}

}
