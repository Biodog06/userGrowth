package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"usergrowth/internal/logs"
	"usergrowth/redis"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		username      string
		password      string
		rawBody       string
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
				m.On("SetCache",
					mock.Anything,
					"123",
					time.Duration(0),
					mock.Anything,
				).Return(nil)
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
			expectMessage: "check password fail",
			expectStatus:  http.StatusUnauthorized,
		},
		{
			name:     "user not found",
			username: "not_exist",
			password: "123",
			mockFindUser: func(m *MockUserRepository) {
				m.On("FindUserByUsername", "not_exist").Return(nil, ErrUserNotFound)
			},
			expectMessage: "invalid username or password",
			expectStatus:  http.StatusUnauthorized,
		},
		{
			name:          "invalid json syntax",
			rawBody:       `"username": "Alice", "password": "123"`,
			expectMessage: "invalid request",
			expectStatus:  http.StatusBadRequest,
		},
		{
			name:          "missing username",
			username:      "",
			password:      "123456",
			expectMessage: "username and password required",
			expectStatus:  http.StatusBadRequest,
		},
		{
			name:          "missing password",
			username:      "Alice",
			password:      "",
			expectMessage: "username and password required",
			expectStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			// 如果不判断 nil 当循环运行到一些用例时 mockRepo 可能是空的，代码会执行空函数
			if tt.mockFindUser != nil {
				tt.mockFindUser(mockRepo)
			}
			mockRDB := new(redis.MockRedis)
			if tt.mockRedis != nil {
				tt.mockRedis(mockRDB)
			}
			logger := logs.NewNopLogger() // 不启用日志

			handler := Login(mockRDB, mockRepo, logger)

			var body []byte
			if tt.rawBody != "" {
				body = []byte(tt.rawBody)
			} else {
				reqBody := map[string]string{
					"username": tt.username,
					"password": tt.password,
				}
				var err error
				body, err = json.Marshal(reqBody)
				if err != nil {
					t.Error(err)
					return
				}
			}

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// 在 gin 中测试
			router := gin.New()
			router.POST("/login", handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectMessage)
			mockRepo.AssertExpectations(t)
		})
	}
}
