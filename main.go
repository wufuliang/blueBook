package main

import (
	"blueBook/internal/repository"
	"blueBook/internal/repository/dao"
	"blueBook/internal/service"
	"blueBook/internal/web"
	"blueBook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	db := initDB()
	server := initWebServer()
	initUserHdl(db, server)

	server.Run(":8080")
}
func initUserHdl(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	hdl := web.NewUserHandler(us)
	hdl.RegisterRoutes(server)
}
func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/blue_book"))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,

		AllowHeaders: []string{"Content-Type", "Authorization"},
		//AllowHeaders: []string{"content-type"},
		// 这个是允许前端访问你的后端响应中带的头部
		ExposeHeaders: []string{"x-jwt-token"},
		//AllowMethods: []string{"POST"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				//if strings.Contains(origin, "localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	//存储数据的，也就是你 userId 存哪里
	//直接存 cookie
	//store := memstore.NewStore([]byte("OwBDs1l7KrjQSa6B9qutsedTy1KiS751"), []byte("aFh0a4WYynAopPMqHoX8Xfu4pYJTrRh2"))
	store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("OwBDs1l7KrjQSa6B9qutsedTy1KiS751"), []byte("aFh0a4WYynAopPMqHoX8Xfu4pYJTrRh2"))
	if err != nil {
		panic(err)
	}
	server.Use(sessions.Sessions("ssid", store), middleware.NewLoginMiddlewareBuilder().Build())
	return server
}
