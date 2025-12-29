package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	config "usergrowth/configs"
	"usergrowth/internal/logs"
	"usergrowth/internal/user"
	"usergrowth/middleware"
	"usergrowth/mysql"
	"usergrowth/redis"

	"github.com/gogf/gf/v2/frame/g"
	//"github.com/gogf/gf/v2/net/ghttp"
	//
	//"go.uber.org/zap"
)

type testUser struct {
	Name string
	Pass string
}
type CustomResponse struct {
	Code int      `json:"code"`
	Msg  string   `json:"message"`
	Data testUser `json:"data"`
}

func main() {
	c := config.NewConfig()
	configPath := os.Getenv("configPath")
	fmt.Println("Config Path:", configPath)
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	c.LoadConfigWithReflex(configPath)

	redisCtx := context.Background()
	rdb := redis.NewRedis(c, redisCtx)
	defer func(rdb redis.Cache) {
		err := rdb.Close()
		if err != nil {
			fmt.Println("not close:", err)
		}
	}(rdb)

	msq := mysql.NewDB(c)
	es := logs.NewEsClient(c)
	middleware.InitJWT(c)

	userLogger := logs.InitLoggerWithES("logs/user.log", c, es)
	defer func(userLogger *logs.MyLogger) {
		err := userLogger.Sync()
		if err != nil {
			fmt.Println("not sync:", err)
		}
	}(userLogger)

	s := g.Server()
	repo := user.NewUserRepository(msq.DB)
	s.SetServerRoot("./static")
	//s.AddStaticPath("/eslog.html", "./static/eslog.html")
	s.BindHandler("/user/register", user.Register(repo, userLogger))
	s.BindHandler("/user/login", user.Login(rdb, repo, userLogger))
	//r.GET("/api/authcheck", middleware.JWTMiddleware(rdb), func(ctx *gin.Context) {
	//	userLogger.RecordInfoLog("check auth on /api/authcheck", zap.String("username", ctx.PostForm("username")))
	//	ctx.JSON(http.StatusOK, gin.H{
	//		"code": 200,
	//		"msg":  "authenticated",
	//	})
	//})
	//s.BindHandler("/api/eslog", logs.GetLogs(es))
	port, err := strconv.Atoi(c.App.Port)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.SetPort(port)
	s.Run()
}
