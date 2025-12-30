package user

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"usergrowth/internal/logs"

	"github.com/gogf/gf/v2/frame/g"
)

type RegisterReq struct {
	g.Meta   `path:"/user/register" method:"post"`
	Username string `json:"username" v:"required#username and password required"`
	Password string `json:"password" v:"required#password and password required"`
}
type RegisterRes struct {
	Name string `json:"name"`
	Pass string `json:"pass"`
}

type Register struct {
	repo       UserRepository
	userLogger logs.Logger
}

func NewRegister(repo UserRepository, logger logs.Logger) *Register {
	return &Register{repo, logger}
}

func (params Register) Register(ctx context.Context, req *RegisterReq) (res *RegisterRes, err error) {

	r := g.RequestFromCtx(ctx)

	sum := md5.Sum([]byte(req.Password))
	hashPass := hex.EncodeToString(sum[:])
	user := &Users{
		Username: req.Username,
		Password: hashPass,
	}

	if err = params.repo.CreateUser(user); err != nil {
		if errors.Is(err, ErrDuplicateUser) {
			params.userLogger.Info(ctx, "duplicate registration", req.Username)

			r.Response.WriteJson(g.Map{
				"code": 400,     // 业务错误码
				"msg":  "用户已存在", // 错误提示
				"data": nil,
			})

			return nil, nil
		}

		// 系统错误（交给中间件处理）
		params.userLogger.Info(ctx, "register db error", req.Username)
		return nil, err
	}

	params.userLogger.Info(ctx, "register success", req.Username)

	r.Response.WriteJson(g.Map{
		"code": 200,
		"msg":  "register success",
		"data": g.Map{
			"name": req.Username,
			"pass": hashPass,
		},
	})

	return nil, nil
}
