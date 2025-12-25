package user

import (
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

var ErrDuplicateUser = errors.New("user already exists")

type User struct {
	userid   string `gorm:"primary_key;auto_increment"`
	username string `gorm:"type:varchar(255);not null;unique"`
	password string `gorm:"type:varchar(255);not null"`
}

type UserRepository interface {
	CreateUser(user *User) error
	FindUserByUsername(username string) (*User, error)
}
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (repo *userRepository) CreateUser(user *User) error {

	if err := repo.db.Create(user).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 { // Error 1062: Duplicate entry
				return ErrDuplicateUser
			}
		}
	} else {
		fmt.Println(err)
		return err
	}
	return nil
}

func (repo *userRepository) FindUserByUsername(username string) (*User, error) {
	var user User
	err := repo.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound // ← 返回自定义错误
		}
		return nil, err // 其他 DB 错误
	}

	return &user, nil
}
