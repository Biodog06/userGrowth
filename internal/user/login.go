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

	"github.com/go-playground/validator/v10"
	"github.com/gogf/gf/v2/net/ghttp"
	"go.uber.org/zap"
)

func Login(rdb redis.Cache, repo UserRepository, userLogger *logs.MyLogger) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		var req struct {
			Username string `json:"username" v:"required#username and password required"`
			Password string `json:"password" v:"required#username and password required"`
		}
		if err := r.Parse(&req); err != nil {
			var errV validator.ValidationErrors
			if errors.As(err, &errV) {
				userLogger.RecordInfoLog("login failed: missing params")
				r.Response.Status = http.StatusBadRequest
				r.Response.WriteJson(ghttp.DefaultHandlerResponse{
					Code:    400,
					Message: "username and password required",
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

		md5Password := md5.Sum([]byte(req.Password))
		hashPass := hex.EncodeToString(md5Password[:])

		//MYSQL VERSION
		if user, err := repo.FindUserByUsername(req.Username); err == nil {
			if user.Password == hashPass {
				token, errM := middleware.GenerateToken(strconv.Itoa(int(user.UserID)))
				if errM != nil {
					fmt.Println(errM)
					return
				}
				fmt.Println(token)
				r.Cookie.SetHttpCookie(&http.Cookie{
					Name:     "jwt-token",
					Value:    token,
					Path:     "/",
					MaxAge:   15 * 60,
					HttpOnly: true,
					Secure:   false,
				})
				if err = rdb.SetCache(token, strconv.Itoa(int(user.UserID)), redis.JWTExpireTime, r.Context()); err != nil {
					fmt.Println(err)
					return
				}
				userLogger.Log(zap.InfoLevel, "login success", zap.String("username", req.Username))
				data := make(map[string]string)
				data["Name"] = req.Username
				data["Pass"] = req.Password
				r.Response.Status = http.StatusOK
				r.Response.WriteJson(
					ghttp.DefaultHandlerResponse{
						Code:    200,
						Message: req.Username + " login success",
						Data:    data,
					})
			} else {
				userLogger.RecordInfoLog("login failed: check user or password failed", zap.String("username", req.Username), zap.String("password", req.Password))
				r.Response.Status = http.StatusUnauthorized
				r.Response.WriteJson(
					ghttp.DefaultHandlerResponse{
						Code:    401,
						Message: "check user or password failed",
					})
			}
		} else {
			if errors.Is(err, ErrUserNotFound) {
				userLogger.RecordInfoLog("login failed: invalid username", zap.String("username", req.Username), zap.String("password", req.Password))
				r.Response.Status = http.StatusUnauthorized
				r.Response.WriteJson(
					ghttp.DefaultHandlerResponse{
						Code:    401,
						Message: "invalid username",
					})
				return
			}
			fmt.Println(err)
			return
		}
	}
}
