package web

import (
	"blueBook/internal/domain"
	"blueBook/internal/service"
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"unicode/utf8"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	nickNameLength       = 6
	profileLength        = 128
	birthdayPattern      = `^\d{4}-\d{2}-\d{2}$`
)

type UserHandler struct {
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	birthdayRexExp *regexp.Regexp

	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		birthdayRexExp: regexp.MustCompile(birthdayPattern, regexp.None),
		svc:            svc,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	// REST 风格
	//server.POST("/user", h.SignUp)
	//server.PUT("/user", h.SignUp)
	//server.GET("/users/:username", h.Profile)
	ug := server.Group("/users")
	// POST /users/signup
	ug.POST("/signup", h.SignUp)
	// POST /users/login
	ug.POST("/login", h.Login)
	// POST /users/edit
	ug.POST("/edit", h.Edit)
	// GET /users/profile
	//ug.GET("/profile", h.Profile)
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不对")
		return
	}

	isPassword, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}

	err = h.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱冲突，请换一个")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := h.svc.Login(ctx, req.Email, req.Password)
	//if errors.Is(err, service.ErrInvalidUserOrPassword) {
	//	ctx.String(http.StatusOK, "用户名或密码错误")
	//	return
	//}
	//sess := sessions.Default(ctx)
	//sess.Set("userId", user.Id)
	//sess.Save()
	switch {
	case err == nil:
		uc := UserClaims{
			Uid:       user.Id,
			UserAgent: ctx.GetHeader("User-Agent"),
			RegisteredClaims: jwt.RegisteredClaims{
				// 1 分钟过期
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
		tokenStr, err := token.SignedString(JWTKey)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
		}
		ctx.Header("x-jwt-token", tokenStr)
		ctx.String(http.StatusOK, "登录成功")
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

	ctx.String(http.StatusOK, "登录成功")
}
func (h *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		NickName string `json:"nickname"`
		Birthday string `json:"birthday"`
		Profile  string `json:"profile"`
	}

	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	if utf8.RuneCountInString(req.NickName) > nickNameLength {
		ctx.String(http.StatusOK, "昵称长度不能超过%d个字符", nickNameLength)
		return
	}
	if utf8.RuneCountInString(req.Profile) > profileLength {
		ctx.String(http.StatusOK, "个人简介长度不能超过%d个字符", profileLength)
		return
	}

	isBirthday, err := h.birthdayRexExp.MatchString(req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isBirthday {
		ctx.String(http.StatusOK, "请输入正确的生日格式")
		return
	}
	sess := sessions.Default(ctx)
	userId := sess.Get("userId")
	err = h.svc.Edit(ctx, domain.User{
		Id:       userId.(int64),
		Birthday: req.Birthday,
		NickName: req.NickName,
		Profile:  req.Profile,
	})
	if errors.Is(err, service.ErrUserNotFound) {
		ctx.String(http.StatusOK, "用户不存在")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
}

var JWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
