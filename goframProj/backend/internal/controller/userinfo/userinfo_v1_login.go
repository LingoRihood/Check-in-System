package userinfo

import (
	"context"

	// "github.com/gogf/gf/v2/errors/gcode"
	// "github.com/gogf/gf/v2/errors/gerror"

	v1 "backend/api/userinfo/v1"
	"backend/internal/model"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	input := &model.LoginInput{
		Username: req.Username,
		Password: req.Password,
	}

	output, err := c.svc.Login(ctx, input)
	if err != nil {
		return nil, gerror.New("用户名或密码错误")
	}

	// 登录成功，更新token
	return &v1.LoginRes{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
	}, err
	// return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
