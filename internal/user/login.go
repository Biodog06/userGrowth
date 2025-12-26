package user

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"usergrowth/internal/logs"
	"usergrowth/middleware"
	"usergrowth/mysql"
	"usergrowth/redis"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Login(rdb *redis.MyRedis, msq *mysql.MyDB, userLogger *logs.MyLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		md5Password := md5.Sum([]byte(password))
		hashPass := hex.EncodeToString(md5Password[:])

		//MYSQL VERSION
		repo := NewUserRepository(msq.DB)
		if user, err := repo.FindUserByUsername(username); err == nil {
			if user.Password == hashPass {
				token, err := middleware.GenerateToken(strconv.Itoa(int(user.UserID)))
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(token)
				ctx.SetCookie("jwt-token", token, 15*60, "/", "", false, true)
				if err := rdb.SetCache(token, strconv.Itoa(int(user.UserID)), redis.JWTExpireTime, ctx); err != nil {
					fmt.Println(err)
					return
				}
				userLogger.Log(zap.InfoLevel, "login success", zap.String("username", username))
				ctx.JSON(http.StatusOK, gin.H{
					"message": username + " login success",
					"data": gin.H{
						"Name": username,
						"Pass": password,
					},
					"code": 200,
				})
			} else {
				userLogger.RecordInfoLog("login failed", zap.String("username", username), zap.String("password", password))
				ctx.JSON(http.StatusOK, gin.H{
					"message": "check password fail",
				})
			}
		} else {
			if errors.Is(err, ErrUserNotFound) {
				userLogger.RecordInfoLog("login failed", zap.String("username", username), zap.String("password", password))
				ctx.JSON(http.StatusOK, gin.H{
					"message": "invalid username or password",
				})
				return
			}

			fmt.Println(err)
			return
		}
	}
}
