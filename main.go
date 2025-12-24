package main

import (
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
	"usergrowth/internal/user"

	"github.com/gin-gonic/gin"
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
	day2_check()
	//day3_check()
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
	testChan := make(chan testUser, 60)
	for i := 0; i < 60; i++ {
		test := testUser{
			Name: strconv.Itoa(rand.Int()),
			Pass: strconv.Itoa(rand.Int()),
		}
		testChan <- test
	}
	r := gin.Default()
	r.POST("/user/register", user.Register)
	r.POST("/user/login", user.Login)

	var wg sync.WaitGroup
	wg.Add(60)
	for i := 0; i < 60; i++ {
		go func() {
			defer wg.Done()
			time.Sleep(1 * time.Second)
			user := <-testChan
			data := url.Values{}
			data.Set("username", user.Name)
			data.Set("password", user.Pass)
			resp, _ := http.Post("http://localhost:8080/user/register",
				"application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
			defer resp.Body.Close()
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
			defer respLogin.Body.Close()
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
	r.Run(":8080")
	//wg.Wait()
}
