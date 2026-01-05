package user

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"usergrowth/internal/logs"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
)

type RegisterReq struct {
	g.Meta   `path:"/user/register" method:"post"`
	Username string `json:"username" v:"required#username and password required"`
	Password string `json:"password" v:"required#password and password required"`
}
type RegisterRes struct {
}

type Register struct {
	repo       UserRepository
	userLogger logs.Logger
}

func NewRegister(repo UserRepository, logger logs.Logger) *Register {
	return &Register{repo, logger}
}

func (params Register) Register(ctx context.Context, req *RegisterReq) (res *RegisterRes, err error) {

	ctx, span := gtrace.NewSpan(ctx, "Register")
	defer span.End()

	r := g.RequestFromCtx(ctx)

	sum := md5.Sum([]byte(req.Password))
	hashPass := hex.EncodeToString(sum[:])
	user := &Users{
		Username: req.Username,
		Password: hashPass,
	}

	if err = params.repo.CreateUser(user); err != nil {
		if errors.Is(err, ErrDuplicateUser) {
			return nil, gerror.NewCode(gcode.CodeValidationFailed, "用户已存在")
		}

		// 系统错误（交给中间件处理）
		params.userLogger.Info(ctx, "register db error", req.Username)
		return nil, err
	}

	params.userLogger.Info(ctx, "register success", req.Username)

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "register success",
		"data": g.Map{
			"name": req.Username,
			"pass": hashPass,
		},
	})

	return nil, nil
}
