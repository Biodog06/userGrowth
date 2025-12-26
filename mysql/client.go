package mysql

import (
	"fmt"
	"strconv"
	"time"
	config "usergrowth/configs"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MyDB struct {
	*gorm.DB
}

func NewDB(cfg *config.Config) *MyDB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", cfg.MySQL.User, cfg.MySQL.Pass, cfg.MySQL.Host, strconv.Itoa(cfg.MySQL.Port), cfg.MySQL.DB)
	fmt.Println(dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return &MyDB{db}
}
