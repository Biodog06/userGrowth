package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	config "usergrowth/configs"
	"usergrowth/internal/logs"
	"usergrowth/internal/observability"
	"usergrowth/internal/user"
	"usergrowth/middleware"
	"usergrowth/mysql"
	"usergrowth/redis"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
	cfg := config.NewConfigManager()
	configPath := os.Getenv("configPath")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	fmt.Println("Config Path:", configPath)
	cfg.LoadConfigWithReflex(configPath)
	cfg.StartWatcher(configPath)
	redisCtx := context.Background()
	rdb := redis.NewRedis(cfg.Config, redisCtx)
	defer func(rdb redis.Cache) {
		err := rdb.Close()
		if err != nil {
			fmt.Println("not close:", err)
		}
	}(rdb)

	msq := mysql.NewDB(cfg.Config)

	middleware.InitJWT(cfg.Config)

	userLogger := logs.NewUserLogger(cfg.Config.App.LogPath)
	errorLogger := logs.NewErrorLogger(cfg.Config.App.LogPath)
	shutdown := observability.InitTracer(cfg.Config.Tracing.ServiceName, cfg.Config.Tracing.Endpoint, cfg.Config.Tracing.Path, errorLogger)
	defer shutdown()
	s := g.Server()

	repo := user.NewUserRepository(msq.DB)
	s.SetServerRoot("./static")
	registerController := user.NewRegister(repo, userLogger)
	loginController := user.NewLogin(rdb, repo, userLogger)
	errorManager := middleware.NewErrorManager(cfg.Config.App.LogPath, &cfg.Config.Middleware, errorLogger)
	loggerManager := middleware.NewLoggerManager(cfg.Config.App.LogPath, &cfg.Config.Middleware)
	jwtManager := middleware.NewJWTManager(rdb, userLogger, &cfg.Config.Middleware)
	traceHandler := middleware.Trace
	esController := logs.NewEsController(cfg.Config)
	authController := user.NewAuthController()
	panicController := user.NewPanicController()

	s.Use(traceHandler, errorManager.ErrorHandler, loggerManager.AccessHandler)

	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(registerController)
		group.Bind(loginController)
		group.Bind(panicController)
	})
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(jwtManager.JWTHandler)
		group.Bind(esController)
		group.Bind(authController)
	})
	port, err := strconv.Atoi(cfg.Config.App.Port)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.SetPort(port)
	s.Run()
}
