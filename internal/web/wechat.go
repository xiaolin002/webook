package web

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"project/internal/service"
	"project/internal/service/oauth2/wechat"
	jwt2 "project/internal/web/jwt"
	"project/pkg/ginx"
)

/**
 * @Description
 * @Date 2025/4/1 20:50
 **/

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	jwt2.Handler
	key             string //这个key是用来生成state的
	stateCookieName string
}

func NewOAuth2WechatHandler(svc wechat.Service,
	userSvc service.UserService, hdl jwt2.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userSvc,
		key:             "",
		stateCookieName: "jwt-state",
		Handler:         hdl,
	}
}

func (o *OAuth2WechatHandler) RegisterRoute(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", o.Auth2URL)  //构建url
	g.Any("/callback", o.Callback) //用于微信回调
}

func (o *OAuth2WechatHandler) Auth2URL(ctx *gin.Context) {
	state := uuid.New()
	url, err := o.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "构建url失败",
			Code: 5,
		})
		return
	}
	err = o.SetStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "服务器异常",
			Code: 5,
		})
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: url,
	})

}

func (o *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := o.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "非法请求",
			Code: 4,
		})
		return
	}

	code := ctx.Query("code")
	wechatInfo, err := o.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "授权码有误",
			Code: 4,
		})
		return
	}
	u, err := o.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		})
		return
	}
	err = o.SetloginToken(ctx, u.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
	return

}
func (o *OAuth2WechatHandler) SetStateCookie(ctx *gin.Context, state string) error {
	claims := StateClaims{
		State: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	// 生成token
	tokenstr, err := token.SignedString([]byte(o.key))
	if err != nil {
		// 这里可以考虑返回一个错误
		return err
	}
	ctx.SetCookie(o.stateCookieName, tokenstr, 600, "/oauth2/wechat/callback", "", false, true)
	return nil
}
func (o *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	// 这里的state是微信请求中带回来的
	state := ctx.Query("state")
	// 这个state 是访问之前设置到cookie中的，需要解析
	tokenstr, err := ctx.Cookie("jwt-state")
	if err != nil {
		return err
	}
	var claims StateClaims
	_, err = jwt.ParseWithClaims(tokenstr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(o.key), nil
	})
	if err != nil {
		return err
	}
	// 验证state
	if claims.State != state {
		return errors.New("state不匹配")
	}
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}
