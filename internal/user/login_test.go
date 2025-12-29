package user

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
	"usergrowth/internal/logs"
	"usergrowth/redis"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/test/gtest" // 引入 gtest
	"github.com/gogf/gf/v2/util/guid"
	"github.com/stretchr/testify/mock"
)

func TestLogin_GTest(t *testing.T) {
	// gtest.C 是 GoFrame 测试的入口
	gtest.C(t, func(t *gtest.T) {
		tests := []struct {
			name          string
			username      string
			password      string
			rawBody       string // 如果需要测试非标准 JSON
			mockFindUser  func(*MockUserRepository)
			mockRedis     func(*redis.MockRedis)
			expectMessage string
			expectStatus  int
		}{
			{
				name:     "login success",
				username: "Alice",
				password: "123456",
				mockFindUser: func(m *MockUserRepository) {
					user := &Users{
						UserID:   123,
						Username: "Alice",
						Password: "e10adc3949ba59abbe56e057f20f883e", // MD5("123456")
					}
					m.On("FindUserByUsername", "Alice").Return(user, nil)
				},
				mockRedis: func(m *redis.MockRedis) {
					m.On("SetCache", mock.Anything, "123", time.Duration(0), mock.Anything).Return(nil)
				},
				expectMessage: "Alice login success",
				expectStatus:  http.StatusOK,
			},
			{
				name:     "password wrong",
				username: "Alice",
				password: "wrong",
				mockFindUser: func(m *MockUserRepository) {
					user := &Users{
						UserID:   123,
						Username: "Alice",
						Password: "e10adc3949ba59abbe56e057f20f883e",
					}
					m.On("FindUserByUsername", "Alice").Return(user, nil)
				},
				expectMessage: "check user or password failed",
				expectStatus:  http.StatusUnauthorized,
			},
			{
				name:     "user not found",
				username: "not_exist",
				password: "123",
				mockFindUser: func(m *MockUserRepository) {
					m.On("FindUserByUsername", "not_exist").Return(nil, ErrUserNotFound)
				},
				expectMessage: "invalid username",
				expectStatus:  http.StatusUnauthorized,
			},
			//{
			//	name:          "missing params",
			//	username:      "",
			//	password:      "",
			//	expectMessage: "username and password required",
			//	expectStatus:  http.StatusBadRequest,
			//},
		}

		for _, tt := range tests {
			// 在 gtest.C 内部，我们通常不使用 t.Run，而是直接执行逻辑
			// 但为了保持结构清晰，我们可以使用普通的块作用域

			// 1. 初始化 Mock
			mockRepo := new(MockUserRepository)
			if tt.mockFindUser != nil {
				tt.mockFindUser(mockRepo)
			}
			mockRDB := new(redis.MockRedis)
			if tt.mockRedis != nil {
				tt.mockRedis(mockRDB)
			}
			logger := logs.NewNopLogger()

			// 2. 初始化 Server (使用 guid 防止冲突)
			s := g.Server(guid.S())
			s.SetLogger(g.Log())
			s.SetAccessLogEnabled(false)
			//s.SetErrorLogEnabled(false)
			s.SetPort(0) // 随机端口

			// 绑定 Handler
			s.BindHandler("/login", Login(mockRDB, mockRepo, logger))

			s.Start()
			// 确保 Loop 结束或出错时关闭 Server
			defer s.Shutdown()

			// 3. 使用 GoFrame 的 Client 发起请求
			client := g.Client()
			client.SetPrefix(fmt.Sprintf("http://127.0.0.1:%d", s.GetListenedPort()))
			client.SetHeader("Content-Type", "application/json")

			var resp *gclient.Response
			var err error

			// 处理请求体
			if tt.rawBody != "" {
				resp, err = client.Post(context.Background(), "/login", tt.rawBody)
			} else {
				// g.Client 自动处理 JSON 序列化
				data := map[string]string{
					"username": tt.username,
					"password": tt.password,
				}
				resp, err = client.ContentJson().Post(context.Background(), "/login", data)
			}

			// 4. gtest 断言风格
			t.Assert(err, nil) // 确保网络请求本身没有错误
			if resp != nil {
				defer resp.Close()
				t.Assert(resp.StatusCode, tt.expectStatus)         // 断言状态码
				t.AssertIN(tt.expectMessage, resp.ReadAllString()) // 断言 Body 包含字符串
			}

			// 验证 Mock 期望
			mockRepo.AssertExpectations(t)
			mockRDB.AssertExpectations(t)
		}
	})
}
