// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMonthlyBonusLog is the golang structure for table user_monthly_bonus_log.
type UserMonthlyBonusLog struct {
	Id          uint64      `json:"id"          orm:"id"          description:"主键ID"`                                     // 主键ID
	UserId      uint64      `json:"userId"      orm:"user_id"     description:"用户ID"`                                     // 用户ID
	YearMonth   string      `json:"yearMonth"   orm:"year_month"  description:"年月 (YYYYMM)"`                              // 年月 (YYYYMM)
	BonusType   int         `json:"bonusType"   orm:"bonus_type"  description:"奖励类型 (1:连续签到3天 2:连续签到7天 3:连续签到15天 4:月满签)"` // 奖励类型 (1:连续签到3天 2:连续签到7天 3:连续签到15天 4:月满签)
	Description string      `json:"description" orm:"description" description:"积分变动说明"`                                   // 积分变动说明
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"  description:"创建时间"`                                     // 创建时间
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"  description:"更新时间"`                                     // 更新时间
	DeletedAt   *gtime.Time `json:"deletedAt"   orm:"deleted_at"  description:"删除时间"`                                     // 删除时间
}
