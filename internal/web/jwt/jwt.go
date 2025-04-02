package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
	"time"
)

/**
 * @Description
 * @Date 2025/4/1 21:28
 **/

type JwtHandler struct {
	// 将才用的加密方法默认进来
	signingMethod jwt.SigningMethod
	rcExpiration  time.Duration
	Client        redis.Cmdable
}

func NewRedisJWTHandler(client redis.Cmdable) Handler {
	return &JwtHandler{
		signingMethod: jwt.SigningMethodHS512,
		rcExpiration:  time.Hour * 24 * 7,
		Client:        client,
	}
}

var _ Handler = &JwtHandler{}

func (j *JwtHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := j.Client.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("token 无效")
	}
	return nil
}

func (j *JwtHandler) SetloginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := j.SetRefreshToken(ctx, uid, ssid)
	if err != nil {
		ctx.JSON(http.StatusOK, "长token设置失败")
		return err
	}

	err = j.SetJwtToken(ctx, uid, ssid)
	if err != nil {
		ctx.JSON(http.StatusOK, "短token设置失败")
		return err
	}
	return nil
}

func (j *JwtHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user").(UserClaims)
	return j.Client.Set(ctx, fmt.Sprintf("users:ssid:%s", uc.Ssid), "", j.rcExpiration).Err()

}

func (j *JwtHandler) ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return authCode
	}
	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}

// 设置短token的  没有长短token之前 都是采用的这个方法
func (j *JwtHandler) SetJwtToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := UserClaims{
		Ssid: ssid,
		Uid:  uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
		},
		UserAgent: ctx.GetHeader("User-Agent"),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES512, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		ctx.JSON(http.StatusOK, "短token设置失败")
		return err
	}
	// 需要注意的是这里要在跨域的处理中将x-jwt-token暴露给前端，将token带过去，同时在AllowHeaders中添加Authorization ,是前端将数据带回
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

var JWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")

type UserClaims struct {
	Uid int64
	jwt.RegisteredClaims
	UserAgent string
	Ssid      string
}

// 这个方法时用来设置refreshToken的 长token
func (h *JwtHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		Ssid: ssid,
		Uid:  uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)),
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, rc)
	tokenStr, err := token.SignedString(RCJWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

var RCJWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgA")

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}
