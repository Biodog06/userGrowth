package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
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
	c.LoadConfig(configPath)

	redisCtx := context.Background()
	rdb := redis.NewRedis(c, redisCtx)
	defer func(rdb *redis.MyRedis) {
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
	day3_check()
	r := gin.Default()
	r.POST("/user/register", user.Register(msq, userLogger))
	r.POST("/user/login", user.Login(rdb, msq, userLogger))
	r.GET("/authcheck", middleware.JWTMiddleware(rdb), func(ctx *gin.Context) {
		userLogger.RecordInfoLog("check auth on /authcheck", zap.String("username", ctx.PostForm("username")))
		ctx.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "authenticated",
		})
	})
	r.GET("/eslog", logs.GetLogs(es))
	err := r.Run(":8080")
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
			data := url.Values{}
			data.Set("username", randomUser.Name)
			data.Set("password", randomUser.Pass)
			resp, err := http.Post("http://localhost:8080/user/register",
				"application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
			defer func(Body io.ReadCloser) {
				err = Body.Close()
				if err != nil {
					fmt.Println("not close:", err)
				}
			}(resp.Body)
			// 打印状态码 (int) 和 状态描述 (string)
			// 1. 读取 Body 的字节数据
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("读取内容失败:", err)
				return
			}

			// 2. 解析 JSON
			var res CustomResponse
			err = json.Unmarshal(bodyBytes, &res)
			if err != nil {
				fmt.Println("解析 JSON 失败:", err)
				// 如果解析失败，打印原始字符串看看服务器到底返回了什么
				fmt.Println("原始返回:", string(bodyBytes))
				return
			}

			// 3. 拿到 message
			fmt.Printf("用户: %s | 密码： %s | 状态码: %d | 消息: %s\n", res.Data.Name, res.Data.Pass, res.Code, res.Msg)

			respLogin, _ := http.Post("http://localhost:8080/user/login",
				"application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					fmt.Println("not close:", err)
				}
			}(respLogin.Body)
			// 打印状态码 (int) 和 状态描述 (string)
			// 1. 读取 Body 的字节数据
			bodyLoginBytes, err := io.ReadAll(respLogin.Body)
			if err != nil {
				fmt.Println("读取内容失败:", err)
				return
			}

			// 2. 解析 JSON
			var resLogin CustomResponse
			err = json.Unmarshal(bodyLoginBytes, &resLogin)
			if err != nil {
				fmt.Println("解析 JSON 失败:", err)
				// 如果解析失败，打印原始字符串看看服务器到底返回了什么
				fmt.Println("原始返回:", string(bodyBytes))
				return
			}

			// 3. 拿到 message
			fmt.Printf("用户: %s | 密码： %s | 状态码: %d | 消息: %s\n", resLogin.Data.Name, resLogin.Data.Pass, resLogin.Code, resLogin.Msg)
		}()
	}
}
