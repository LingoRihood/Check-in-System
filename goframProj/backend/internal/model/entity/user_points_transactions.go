// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserPointsTransactions is the golang structure for table user_points_transactions.
type UserPointsTransactions struct {
	Id              uint64      `json:"id"              orm:"id"               description:"主键ID"`                                 // 主键ID
	UserId          uint64      `json:"userId"          orm:"user_id"          description:"用户ID"`                                 // 用户ID
	PointsChange    int64       `json:"pointsChange"    orm:"points_change"    description:"积分变动值（正数为增加，负数为扣除）"`                   // 积分变动值（正数为增加，负数为扣除）
	CurrentBalance  int64       `json:"currentBalance"  orm:"current_balance"  description:"当前余额"`                                 // 当前余额
	TransactionType int         `json:"transactionType" orm:"transaction_type" description:"交易类型(1:签到 2:连续签到 3:补签 4:每日任务 5:福利任务)"` // 交易类型(1:签到 2:连续签到 3:补签 4:每日任务 5:福利任务)
	Description     string      `json:"description"     orm:"description"      description:"积分变动说明"`                               // 积分变动说明
	ExtJson         string      `json:"extJson"         orm:"ext_json"         description:"扩展字段"`                                 // 扩展字段
	CreatedAt       *gtime.Time `json:"createdAt"       orm:"created_at"       description:"创建时间"`                                 // 创建时间
	UpdatedAt       *gtime.Time `json:"updatedAt"       orm:"updated_at"       description:"更新时间"`                                 // 更新时间
	DeletedAt       *gtime.Time `json:"deletedAt"       orm:"deleted_at"       description:"删除时间"`                                 // 删除时间
}
