package user

import "github.com/stretchr/testify/mock"

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindUserByUsername(username string) (*Users, error) {
	args := m.Called(username)

	// 增加 nil 检查，否则登陆失败返回 nil 时会 panic
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Users), args.Error(1)
}

func (m *MockUserRepository) CreateUser(user *Users) error {
	args := m.Called(user)
	return args.Error(0)
}
