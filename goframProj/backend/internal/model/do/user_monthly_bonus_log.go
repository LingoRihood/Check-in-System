// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMonthlyBonusLog is the golang structure of table user_monthly_bonus_log for DAO operations like Where/Data.
type UserMonthlyBonusLog struct {
	g.Meta      `orm:"table:user_monthly_bonus_log, do:true"`
	Id          any         // 主键ID
	UserId      any         // 用户ID
	YearMonth   any         // 年月 (YYYYMM)
	BonusType   any         // 奖励类型 (1:连续签到3天 2:连续签到7天 3:连续签到15天 4:月满签)
	Description any         // 积分变动说明
	CreatedAt   *gtime.Time // 创建时间
	UpdatedAt   *gtime.Time // 更新时间
	DeletedAt   *gtime.Time // 删除时间
}
