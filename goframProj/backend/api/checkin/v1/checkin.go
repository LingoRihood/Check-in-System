package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// 定义签到相关的 API
type DailyReq struct {
	g.Meta `path:"/checkins" method:"POST" tags:"签到" summary:"每日签到"`
}

type DailyRes struct{}
