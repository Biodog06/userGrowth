package user

import (
	"context"
	"usergrowth/internal/logs"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type AuthReq struct {
	g.Meta `path:"/api/authcheck" method:"get"`
}

type AuthRes struct {
	Username string `json:"username"`
	UserID   string `json:"userid"`
}

type AuthController struct {
	userLogger logs.Logger
}

func NewAuthController(userLogger logs.Logger) *AuthController {
	return &AuthController{userLogger: userLogger}
}

func (c *AuthController) Check(ctx context.Context, req *AuthReq) (res *AuthRes, err error) {
	r := g.RequestFromCtx(ctx)
	// 从上下文获取用户信息（由 JWT 中间件设置）
	username := r.GetCtxVar("username").String()
	userid := r.GetCtxVar("userid").String()

	c.userLogger.Info(ctx, "check auth success", "username", username, "userid", userid)

	r.Response.WriteJson(ghttp.DefaultHandlerResponse{
		Code:    200,
		Message: "authenticated",
		Data: &AuthRes{
			Username: username,
			UserID:   userid,
		},
	})
	return nil, nil
}
