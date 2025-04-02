package jwt

import "github.com/gin-gonic/gin"

/**
 * @Description
 * @Date 2025/4/2 20:57
 **/

type Handler interface {
	ExtractToken(ctx *gin.Context) string
	SetJwtToken(ctx *gin.Context, uid int64, ssid string) error
	ClearToken(ctx *gin.Context) error
	SetloginToken(ctx *gin.Context, uid int64) error
	CheckSession(ctx *gin.Context, ssid string) error
}
