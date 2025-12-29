package user

import (
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

var ErrDuplicateUser = errors.New("user already exists")

type Users struct {
	UserID   uint   `gorm:"primaryKey;autoIncrement"`               // 自增主键
	Username string `gorm:"type:varchar(255);not null;uniqueIndex"` // 唯一索引
	Password string `gorm:"type:varchar(255);not null"`
}

type UserRepository interface {
	CreateUser(user *Users) error
	FindUserByUsername(username string) (*Users, error)
}
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (repo *userRepository) CreateUser(user *Users) error {

	err := repo.db.AutoMigrate(&Users{})
	if err != nil {
		panic("failed to migrate table")
	}
	if err = repo.db.Create(user).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 { // Error 1062: Duplicate entry
				return ErrDuplicateUser
			}
		} else {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (repo *userRepository) FindUserByUsername(username string) (*Users, error) {
	var user Users
	err := repo.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound // ← 返回自定义错误
		}
		return nil, err // 其他 DB 错误
	}

	return &user, nil
}
