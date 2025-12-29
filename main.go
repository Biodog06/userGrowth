package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	config "usergrowth/configs"
	"usergrowth/internal/logs"
	"usergrowth/internal/user"
	"usergrowth/middleware"
	"usergrowth/mysql"
	"usergrowth/redis"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	//day2_check()
	//day3_check()
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
	//day3_check()
	r := gin.Default()
	repo := user.NewUserRepository(msq.DB)
	r.POST("/user/register", user.Register(repo, userLogger))
	r.POST("/user/login", user.Login(rdb, repo, userLogger))
	r.GET("/authcheck", middleware.JWTMiddleware(rdb), func(ctx *gin.Context) {
		userLogger.RecordInfoLog("check auth on /authcheck", zap.String("username", ctx.PostForm("username")))
		ctx.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "authenticated",
		})
	})
	r.GET("/eslog", logs.GetLogs(es))
	addr := fmt.Sprintf(":%s", c.App.Port)
	err := r.Run(addr)
	if err != nil {
		return
	}
}

func day2_check() {
	c := config.NewConfig()
	configPath := os.Getenv("configPath")
	fmt.Println("Config Path:", configPath)
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	c.LoadConfig(configPath)
	c.PrintConfig()
}

func day3_check() {
	testChan := make(chan testUser, 10)
	for i := 0; i < 10; i++ {
		test := testUser{
			Name: strconv.Itoa(rand.Int()),
			Pass: strconv.Itoa(rand.Int()),
		}
		testChan <- test
	}

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			time.Sleep(1 * time.Second)
			randomUser := <-testChan

			// ========== 注册：改用 JSON ==========
			reqBody := struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}{
				Username: randomUser.Name,
				Password: randomUser.Pass,
			}
			jsonBody, _ := json.Marshal(reqBody)

			resp, err := http.Post(
				"http://localhost:8080/user/register",
				"application/json",        // ✅
				bytes.NewReader(jsonBody), // ✅
			)
			if err != nil {
				fmt.Println("注册请求失败:", err)
				return
			}
			defer func(Body io.ReadCloser) {
				err = Body.Close()
				if err != nil {
					fmt.Println(err)
					return
				}
			}(resp.Body)

			bodyBytes, _ := io.ReadAll(resp.Body)
			var res CustomResponse
			_ = json.Unmarshal(bodyBytes, &res)
			fmt.Printf("[注册] 用户: %s | 密码: %s | 状态码: %d | 消息: %s\n",
				res.Data.Name, res.Data.Pass, res.Code, res.Msg)

			// ========== 登录：改用 JSON ==========
			loginBody, _ := json.Marshal(reqBody) // 复用结构体
			respLogin, err := http.Post(
				"http://localhost:8080/user/login",
				"application/json",         // ✅
				bytes.NewReader(loginBody), // ✅
			)
			if err != nil {
				fmt.Println("登录请求失败:", err)
				return
			}
			defer func(Body io.ReadCloser) {
				err = Body.Close()
				if err != nil {
					fmt.Println(err)
					return
				}
			}(respLogin.Body)

			bodyLoginBytes, _ := io.ReadAll(respLogin.Body)
			var resLogin CustomResponse
			_ = json.Unmarshal(bodyLoginBytes, &resLogin)
			fmt.Printf("[登录] 用户: %s | 密码: %s | 状态码: %d | 消息: %s\n",
				resLogin.Data.Name, resLogin.Data.Pass, resLogin.Code, resLogin.Msg)
		}()
	}
	//wg.Wait()
}
