package user

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type PanicReq struct {
	g.Meta `path:"/api/panictest" method:"get"`
}

type PanicRes struct {
}

type PanicController struct {
}

func NewPanicController() *PanicController {
	return &PanicController{}
}

func (c *PanicController) Panic(ctx context.Context, req *PanicReq) (res *PanicRes, err error) {
	panic("panic! panic!")
}
