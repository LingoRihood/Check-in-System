package checkin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "backend/api/checkin/v1"
)

// Daily 每日签到接口实现
func (c *ControllerV1) Daily(ctx context.Context, req *v1.DailyReq) (res *v1.DailyRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
