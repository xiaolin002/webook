package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"project/internal/domain"
	"project/internal/service"
	jwt2 "project/internal/web/jwt"
	"project/pkg/ginx"
)

/**
 * @Description  用户模块
 * @Date 2024/2/28 11:12
 **/

// 后端处理步骤
// 1.接受请求并校验
// 2.调用业务逻辑处理请求
// 3.根据业务逻辑处理结果返回响应

const (
	emailRegexpPattern    = ""
	passwordRegexpPattern = ``
	bizLogin              = "login"
)

type UsersHandler struct {
	// 组合
	jwt2.Handler
	emailRegexpRex    *regexp.Regexp
	passwordRegexpRex *regexp.Regexp
	svc               service.UserService
	codeSvc           service.CodeService
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, hdl jwt2.Handler) *UsersHandler {
	return &UsersHandler{
		codeSvc:           codeSvc,
		emailRegexpRex:    regexp.MustCompile(emailRegexpPattern, regexp.None),
		passwordRegexpRex: regexp.MustCompile(passwordRegexpPattern, regexp.None),
		svc:               svc,
		Handler:           hdl,
	}

}

func (u *UsersHandler) RegisterRoute(server *gin.Engine) {
	user := server.Group("/user")
	user.POST("/login", u.Login)
	user.POST("/signup", u.SignUp)
	user.GET("/profile", u.Profile)
	user.GET("/refresh_token", u.RefreshToken)
	user.POST("/edit", u.Edit)
	user.POST("/loginjwt", u.LoginJwt)
	user.POST("/login_sms/code/send", u.SendSmsLoginCode)
	user.POST("/login_sms", u.LoginSms)
	// session 登出
	user.POST("/logout", u.SessionLogOut)
	// jwt 登出
	user.POST("/logoutjwt", u.LogoutJwt)
}

// 注册

func (u *UsersHandler) SignUp(ctx *gin.Context) {
	// 这里来说需要一个正则匹配来验证邮箱  其实前端可以做验证
	// 内部类
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword" `
		Password        string `json:"password"`
	}
	var sign SignUpReq

	// 利用bind 中的contentType 来判断采用那种方式与数据进行绑定 错误就会400
	if err := ctx.Bind(&sign); err != nil {
		return
	}

	// 密码是否一致

	if sign.Password != sign.ConfirmPassword {
		ctx.JSON(http.StatusBadRequest, "两次密码不一致")
		return
	}

	//校验邮箱(正则)

	isEmail, err := u.emailRegexpRex.MatchString(sign.Email)
	if err != nil {
		// 系统默认超时
		ctx.JSON(http.StatusBadRequest, "系统错误")
		return
	}
	if !isEmail {
		ctx.JSON(http.StatusBadRequest, "非法邮箱格式")
		return
	}

	//校验密码(正则)

	isPassword, err := u.passwordRegexpRex.MatchString(sign.Password)
	if err != nil {
		//系统默认超时
		ctx.JSON(http.StatusBadRequest, "系统错误")
		return
	}
	if !isPassword {
		ctx.JSON(http.StatusBadRequest, "非法邮箱格式")
		return
	}
	err = u.svc.SignUp(ctx, domain.User{
		Email:    sign.Email,
		Password: sign.Password,
	})

	switch err {
	case nil:
		ctx.JSON(http.StatusOK, "注册成功")
		return
	case service.ErrDuplicateEmail:
		ctx.JSON(http.StatusOK, "邮箱冲突")
		return
	default:
		ctx.JSON(http.StatusOK, "系统错误")
	}

}

// 登录

func (u *UsersHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var login LoginReq
	if err := ctx.Bind(&login); err != nil {
		ctx.JSON(http.StatusOK, "系统错误")
		return
	}
	h, err := u.svc.Login(ctx, login.Email, login.Password)

	switch err {
	case nil:
		session := sessions.Default(ctx)
		session.Set("uid", h.Id)
		session.Options(sessions.Options{
			MaxAge: 900,
		})
		err = session.Save()
		if err != nil {
			log.Println(err)
			ctx.String(http.StatusOK, "服务器异常")
			return
		}
		ctx.JSON(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.JSON(http.StatusOK, "用户名或密码错误")

	default:
		ctx.JSON(http.StatusOK, "系统错误")

	}
}

// 采用jwt来登录

func (u *UsersHandler) LoginJwt(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var login LoginReq
	if err := ctx.Bind(&login); err != nil {
		ctx.JSON(http.StatusOK, "系统错误")
		return
	}
	h, err := u.svc.Login(ctx, login.Email, login.Password)

	switch err {
	case nil:
		err = u.SetloginToken(ctx, h.Id)
		if err != nil {
			ctx.JSON(http.StatusOK, "系统错误")
			return
		}
		ctx.JSON(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.JSON(http.StatusOK, "用户名或密码错误")

	default:
		ctx.JSON(http.StatusOK, "系统错误")

	}

}

// 获取信息

func (u *UsersHandler) Profile(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "获取信息")
}

// 修改

func (u *UsersHandler) Edit(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "修改")
}

func (u *UsersHandler) SendSmsLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, "系统错误")
		return
	}
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "手机号不能为空",
		})
		return
	}
	err := u.codeSvc.Send(ctx, bizLogin, req.Phone)
	switch err {

	case nil:
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 0,
			Msg:  "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "短信发送太频繁,请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 补日志
	}

}

func (u *UsersHandler) LoginSms(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, "系统错误")
		return
	}
	if req.Phone == "" || req.Code == "" {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "手机号或验证码不能为空",
		})
		return
	}
	ok, err := u.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "验证码错误",
		})
		return
	}
	h, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = u.SetloginToken(ctx, h.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Code: 0,
		Msg:  "登录成功",
	})

}

func (u *UsersHandler) RefreshToken(ctx *gin.Context) {
	// 约定前端在Authorization 里边带上refresh_token
	// 除了该接口 其他接口约定前端在Authorization 里边带上access_token
	// 由于是需要过期了  所以需要携带长token
	// 因此可以吧中间件的判断抽出来 成一个方法、
	tokenStr := u.ExtractToken(ctx)
	var rc jwt2.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return jwt2.RCJWTKey, nil
	})
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if token == nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 这里同时要验证一下ssid是否储存在redis中
	err = u.CheckSession(ctx, rc.Ssid)
	if err != nil {

		// 系统错误或者用户已经退出登录了 或者redis崩溃就不需要做校验了
		ctx.AbortWithStatus(http.StatusUnauthorized)

		return
	}

	err = u.SetJwtToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "Ok",
	})

}

// SessionLogOut 如果用session的话 就可以这种方式退出
func (u *UsersHandler) SessionLogOut(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: -1,
	})
	sess.Save()

}

func (u *UsersHandler) LogoutJwt(ctx *gin.Context) {
	err := u.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "退出登录成功",
	})

}
