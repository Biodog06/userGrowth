package user

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"usergrowth/internal/logs"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

//var userMap sync.Map
//
//// 给每个 user 一个用户id，比较方便
//var idMap sync.Map
//
//var id = 1

func Register(repo UserRepository, userLogger *logs.MyLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		var errV validator.ValidationErrors
		if err := ctx.ShouldBindJSON(&req); err != nil {
			if errors.As(err, &errV) {
				userLogger.RecordInfoLog("login failed: missing params")
				ctx.JSON(http.StatusBadRequest, gin.H{
					"message": "username and password required",
					"code":    400,
				})
				return
			}
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid request",
				"code":    400,
			})
			userLogger.RecordInfoLog("login failed: invalid request")
			return
		}
		//MYSQL VERSION
		data := md5.Sum([]byte(req.Password))
		hashPass := hex.EncodeToString(data[:])
		user := &Users{
			Username: req.Username,
			Password: hashPass,
		}
		if err := repo.CreateUser(user); err != nil {
			if errors.Is(err, ErrDuplicateUser) {
				userLogger.RecordInfoLog("duplicate registration", zap.String("username", req.Username), zap.String("password", req.Password))
				ctx.JSON(http.StatusBadRequest, gin.H{
					"message": "user already exists",
					"code":    400,
				})
			} else {
				fmt.Println(err)
				return
			}
		} else {
			userLogger.RecordInfoLog("register success", zap.String("username", req.Username), zap.String("password", req.Password))
			ctx.JSON(http.StatusOK, gin.H{
				"message": "register success",
				"data": gin.H{
					"Name": req.Username,
					"Pass": req.Password,
				},
				"code": 200,
			})
		}

	}
}
