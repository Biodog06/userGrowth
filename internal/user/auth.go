package user

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type AuthReq struct {
	g.Meta `path:"/api/authcheck" method:"get"`
}

type AuthRes struct {
	Username string `json:"username"`
	UserID   string `json:"userid"`
}

type AuthController struct {
}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func (c *AuthController) Check(ctx context.Context, req *AuthReq) (res *AuthRes, err error) {
	r := g.RequestFromCtx(ctx)
	// 从上下文获取用户信息（由 JWT 中间件设置）
	username := r.GetCtxVar("username").String()
	userid := r.GetCtxVar("userid").String()

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "authenticated",
		"data": &AuthRes{
			Username: username,
			UserID:   userid,
		},
	})
	return nil, nil
}
