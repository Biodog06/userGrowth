package user

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"usergrowth/internal/logs"
	"usergrowth/middleware"
	"usergrowth/redis"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"go.opentelemetry.io/otel/attribute"
)

type LoginReq struct {
	g.Meta   `path:"/user/login" method:"post"`
	Username string `json:"username" v:"required#username required"`
	Password string `json:"password" v:"required#password required"`
}

type LoginRes struct {
}

type Login struct {
	rdb        redis.Cache
	repo       UserRepository
	userLogger logs.Logger
}

func NewLogin(rdb redis.Cache, repo UserRepository, logger logs.Logger) *Login {
	return &Login{
		rdb:        rdb,
		repo:       repo,
		userLogger: logger,
	}
}

func (params *Login) Login(ctx context.Context, req *LoginReq) (res *LoginRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "Login")
	defer span.End()

	r := g.RequestFromCtx(ctx)

	md5Password := md5.Sum([]byte(req.Password))
	hashPass := hex.EncodeToString(md5Password[:])

	user, err := params.repo.FindUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, gerror.NewCode(gcode.CodeNotAuthorized, "invalid username or password")
		}

		//params.userLogger.Error(ctx, "login db error", err)
		return nil, err
	}

	if user.Password != hashPass {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "invalid username or password")
	}

	token, err := middleware.GenerateToken(strconv.Itoa(int(user.UserID)))
	if err != nil {
		//params.userLogger.Error(ctx, "generate token error", err)
		return nil, err
	}

	r.Cookie.SetHttpCookie(&http.Cookie{
		Name:     "jwt-token",
		Value:    token,
		Path:     "/",
		MaxAge:   15 * 60,
		HttpOnly: true,
		Secure:   false,
	})

	if err = params.rdb.SetCache(token, strconv.Itoa(int(user.UserID)), redis.JWTExpireTime, ctx); err != nil {
		//params.userLogger.Error(ctx, "redis set cache error", err)
		return nil, err
	}

	span.SetAttributes(attribute.String("user.id", strconv.Itoa(int(user.UserID))))

	params.userLogger.Info(ctx, fmt.Sprintf("login success: %s", req.Username), "userid", user.UserID)

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "login success",
		"data": g.Map{
			"name":  user.Username,
			"token": token,
		},
	})
	return nil, nil
}

type LogoutReq struct {
	g.Meta `path:"/user/logout" method:"post"`
}

type LogoutRes struct {
}

func (params *Login) Logout(ctx context.Context, req *LogoutReq) (res *LogoutRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "Logout")
	defer span.End()

	r := g.RequestFromCtx(ctx)
	tokenString := r.Cookie.Get("jwt-token").String()

	var userId string
	if tokenString == "" {
		params.userLogger.Info(ctx, "Logout: Failed to get cookie or token is empty")
	} else {
		// Try to get userId from Redis before deleting
		val, err := params.rdb.GetCache(tokenString, ctx)
		if err == nil {
			userId = val
		} else {
			claims, err1 := middleware.ValidateToken(tokenString)
			if err1 == nil {
				userId = claims.UserId
			} else {
				params.userLogger.Info(ctx, fmt.Sprintf("Logout: Failed to validate token: %v", err1))
			}
		}

		params.userLogger.Info(ctx, fmt.Sprintf("Logout: Got token from cookie: %s", tokenString))
		err = params.rdb.DeleteCache(tokenString, ctx)
		if err != nil {
			params.userLogger.Info(ctx, fmt.Sprintf("Logout: Failed to delete from redis: %v", err))
		} else {
			params.userLogger.Info(ctx, "Logout: Successfully deleted from redis")
		}
	}

	if userId != "" {
		span.SetAttributes(attribute.String("user.id", userId))
	}

	r.Cookie.SetHttpCookie(&http.Cookie{
		Name:     "jwt-token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
	})

	params.userLogger.Info(ctx, "user logout", "userid", userId)
	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "logout success",
	})
	return nil, nil
}
