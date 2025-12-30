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
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
	cfg := config.NewConfig()
	configPath := os.Getenv("configPath")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	fmt.Println("Config Path:", configPath)
	cfg.LoadConfigWithReflex(configPath)

	redisCtx := context.Background()
	rdb := redis.NewRedis(cfg, redisCtx)
	defer func(rdb redis.Cache) {
		err := rdb.Close()
		if err != nil {
			fmt.Println("not close:", err)
		}
	}(rdb)

	msq := mysql.NewDB(cfg)

	middleware.InitJWT(cfg)

	userLogger := logs.NewUserLogger(cfg.App.LogPath)

	s := g.Server()
	repo := user.NewUserRepository(msq.DB)
	s.SetServerRoot("./static")
	registerParam := user.NewRegister(repo, userLogger)
	loginParam := user.NewLogin(rdb, repo, userLogger)
	errorManager := middleware.NewErrorManager(cfg.App.LogPath)
	loggerManager := middleware.NewLoggerManager(cfg.App.LogPath)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(errorManager.ErrorHandler)
		group.Middleware(loggerManager.AccessHandler)
		group.Bind(registerParam)
		group.Bind(loginParam)
	})
	es := logs.NewEsClient(cfg)
	s.BindHandler("/api/eslog", logs.GetLogs(es))
	//r.GET("/api/authcheck", middleware.JWTMiddleware(rdb), func(ctx *gin.Context) {
	//	userLogger.RecordInfoLog("check auth on /api/authcheck", zap.String("username", ctx.PostForm("username")))
	//	ctx.JSON(http.StatusOK, gin.H{
	//		"code": 200,
	//		"msg":  "authenticated",
	//	})
	//})
	port, err := strconv.Atoi(cfg.App.Port)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.SetPort(port)
	s.Run()
}
