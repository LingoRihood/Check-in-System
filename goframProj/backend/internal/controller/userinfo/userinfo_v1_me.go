package userinfo

import (
	v1 "backend/api/userinfo/v1"
	"backend/internal/consts"
	"context"
	"strconv"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) Me(ctx context.Context, req *v1.MeReq) (res *v1.MeRes, err error) {
	// 从请求上下文获取 userId
	userId, ok := ctx.Value(consts.CtxKeyUserID).(uint64)

	g.Log().Debugf(ctx, "从请求上下文中获取 userId: %d", userId)
	if !ok || userId == 0 {
		return nil, gerror.New("用户信息获取失败")
	}

	// 根据用户id获取用户信息
	// FormatUint 用来将 uint64 类型的整数转换为字符串
	userInfo, err := c.svc.GetInfo(ctx, strconv.FormatUint(userId, 10))

	// userInfo, err := c.svc.GetInfo(ctx, "7848643939821819")

	if err != nil {
		return nil, gerror.New("获取用户信息失败")
	}

	// 返回用户信息
	return &v1.MeRes{
		Username: userInfo.Username,
		Avatar:   userInfo.Avatar,
		Email:    userInfo.Email,
	}, err
}
