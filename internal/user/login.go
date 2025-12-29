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
	"usergrowth/redis"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func Login(rdb redis.Cache, repo UserRepository, userLogger *logs.MyLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			var errV validator.ValidationErrors
			if errors.As(err, &errV) {
				userLogger.RecordInfoLog("login failed: missing params")
				ctx.JSON(http.StatusBadRequest, gin.H{
					"message": "username and password required",
					"code":    400,
				})
				return
			}
			userLogger.RecordInfoLog("login failed: invalid request")
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid request",
				"code":    400,
			})
			return
		}

		md5Password := md5.Sum([]byte(req.Password))
		hashPass := hex.EncodeToString(md5Password[:])

		//MYSQL VERSION
		if user, err := repo.FindUserByUsername(req.Username); err == nil {
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
				userLogger.Log(zap.InfoLevel, "login success", zap.String("username", req.Username))
				ctx.JSON(http.StatusOK, gin.H{
					"message": req.Username + " login success",
					"data": gin.H{
						"Name": req.Username,
						"Pass": req.Password,
					},
					"code": 200,
				})
			} else {
				userLogger.RecordInfoLog("login failed", zap.String("username", req.Username), zap.String("password", req.Password))
				ctx.JSON(http.StatusUnauthorized, gin.H{
					"message": "check password fail",
					"code":    401,
				})
			}
		} else {
			if errors.Is(err, ErrUserNotFound) {
				userLogger.RecordInfoLog("login failed", zap.String("username", req.Username), zap.String("password", req.Password))
				ctx.JSON(http.StatusUnauthorized, gin.H{
					"message": "invalid username or password",
					"code":    401,
				})
				return
			}

			fmt.Println(err)
			return
		}
	}
}
