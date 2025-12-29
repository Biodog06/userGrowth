package redis

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockRedis struct{ mock.Mock }

func (m *MockRedis) SetCache(key, value string, expire time.Duration, ctx context.Context) error {
	return m.Called(key, value, expire, ctx).Error(0)
}

func (m *MockRedis) GetCache(key string, ctx context.Context) (string, error) {
	args := m.Called(key, ctx)
	return args.String(0), args.Error(1)
}

func (m *MockRedis) Close() error {
	return m.Called().Error(0)
}
