package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"usergrowth/internal/logs"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		username       string
		password       string
		mockCreateUser func(repository *MockUserRepository)
		expectMessage  string
		expectStatus   int
	}{
		{
			name:     "register success",
			username: "Alice",
			password: "123456",
			mockCreateUser: func(m *MockUserRepository) {
				user := &Users{
					UserID:   0,
					Username: "Alice",
					Password: "e10adc3949ba59abbe56e057f20f883e",
				}
				m.On("CreateUser", user).Return(nil)
			},
			expectMessage: "register success",
			expectStatus:  http.StatusOK,
		},
		{
			name:     "duplicate registration",
			username: "Alice",
			password: "123456",
			mockCreateUser: func(m *MockUserRepository) {
				existingUser := &Users{
					UserID:   0,
					Username: "Alice",
					Password: "e10adc3949ba59abbe56e057f20f883e",
				}
				m.On("CreateUser", existingUser).Return(ErrDuplicateUser)
			},
			expectMessage: "user already exists",
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

			if tt.mockCreateUser != nil {
				tt.mockCreateUser(mockRepo)
			}
			logger := logs.NewNopLogger()

			handler := Register(mockRepo, logger)

			reqBody := map[string]string{
				"username": tt.username,
				"password": tt.password,
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router := gin.New()
			router.POST("/register", handler)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectMessage)
			mockRepo.AssertExpectations(t)
		})
	}

}
