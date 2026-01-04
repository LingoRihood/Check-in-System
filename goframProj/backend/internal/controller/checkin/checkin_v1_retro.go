package checkin

import (
	"backend/internal/consts"
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	// "github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "backend/api/checkin/v1"
)

func (c *ControllerV1) Retro(ctx context.Context, req *v1.RetroReq) (*v1.RetroRes, error) {
	// 1. 校验日期格式, 2025-07-01
	t, err := time.Parse(time.DateOnly, req.Date)
	if err != nil {
		return nil, gerror.New("日期格式不正确")
	}

	// 从请求上下文中获取 userid
	userId, ok := ctx.Value(consts.CtxKeyUserID).(uint64)
	g.Log().Debugf(ctx, "从请求上下文中获取 userId: %d", userId)
	if !ok || userId == 0 {
		return nil, gerror.New("用户信息获取失败")
	}

	// 2. 调用 service 层补签逻辑
	if err = c.svc.Retro(ctx, userId, t); err != nil {
		return nil, err
	}
	return &v1.RetroRes{}, nil
}
