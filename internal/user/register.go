package user

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"usergrowth/internal/logs"

	"github.com/go-playground/validator/v10"
	"github.com/gogf/gf/v2/net/ghttp"
	"go.uber.org/zap"
)

//var userMap sync.Map
//
//// 给每个 user 一个用户id，比较方便
//var idMap sync.Map
//
//var id = 1

func Register(repo UserRepository, userLogger *logs.MyLogger) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		var req struct {
			Username string `json:"username" v:"required#username and password required"`
			Password string `json:"password" v:"required#username and password required"`
		}
		var errV validator.ValidationErrors
		if err := r.Parse(&req); err != nil {
			if errors.As(err, &errV) {
				userLogger.RecordInfoLog("login failed: missing params")
				r.Response.WriteStatus(http.StatusBadRequest)
				r.Response.WriteJson(ghttp.DefaultHandlerResponse{
					Code:    400,
					Message: "useranme and password required",
				})
				return
			}
			fmt.Println(err)
			userLogger.RecordInfoLog("login failed: unknown error")
			r.Response.WriteStatus(http.StatusBadRequest)
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    400,
				Message: "unknown error",
			})
			return
		}
		//MYSQL VERSION
		sum := md5.Sum([]byte(req.Password))
		hashPass := hex.EncodeToString(sum[:])
		user := &Users{
			Username: req.Username,
			Password: hashPass,
		}
		if err := repo.CreateUser(user); err != nil {
			if errors.Is(err, ErrDuplicateUser) {
				userLogger.RecordInfoLog("duplicate registration", zap.String("username", req.Username), zap.String("password", req.Password))
				r.Response.WriteStatus(http.StatusBadRequest)
				r.Response.WriteJson(ghttp.DefaultHandlerResponse{
					Code:    400,
					Message: "user already exists",
				})
			} else {
				fmt.Println(err)
				return
			}
		} else {
			userLogger.RecordInfoLog("register success", zap.String("username", req.Username), zap.String("password", req.Password))
			data := make(map[string]string)
			data["Name"] = req.Username
			data["Pass"] = req.Password
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    200,
				Message: "register success",
				Data:    data,
			})
		}
	}
}
