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
	"time"
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
)

type UsersHandler struct {
	emailRegexpRex    *regexp.Regexp
	passwordRegexpRex *regexp.Regexp
	svc               *service.UsersService
}

func NewUserHandler(svc *service.UsersService) *UsersHandler {
	return &UsersHandler{
		emailRegexpRex:    regexp.MustCompile(emailRegexpPattern, regexp.None),
		passwordRegexpRex: regexp.MustCompile(passwordRegexpPattern, regexp.None),
		svc:               svc,
	}

}

func (u *UsersHandler) RegisterRouter(server *gin.Engine) {
	user := server.Group("/user")
	user.POST("/login", u.Login)
	user.POST("/signup", u.SignUp)
	user.GET("/profile", u.Profile)
	user.POST("/edit", u.Edit)
	user.POST("/loginjwt", u.LoginJwt)
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
	uc := UserClaims{
		Uid: h.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
		},
		UserAgent: ctx.GetHeader("User-Agent"),
	}
	switch err {
	case nil:
		token := jwt.NewWithClaims(jwt.SigningMethodES512, uc)
		tokenStr, err := token.SignedString(Jwtkey)
		if err != nil {
			ctx.JSON(http.StatusOK, "系统错误")
		}
		// 需要注意的是这里要在跨域的处理中将x-jwt-token暴露给前端，将token带过去，同时在AllowHeaders中添加Authorization ,是前端将数据带回
		ctx.Header("x-jwt-token", tokenStr)

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

var Jwtkey = []byte("")

type UserClaims struct {
	Uid int64
	jwt.RegisteredClaims
	UserAgent string
}
