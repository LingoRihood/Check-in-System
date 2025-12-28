// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserPointsTransactions is the golang structure of table user_points_transactions for DAO operations like Where/Data.
type UserPointsTransactions struct {
	g.Meta          `orm:"table:user_points_transactions, do:true"`
	Id              any         // 主键ID
	UserId          any         // 用户ID
	PointsChange    any         // 积分变动值（正数为增加，负数为扣除）
	CurrentBalance  any         // 当前余额
	TransactionType any         // 交易类型(1:签到 2:连续签到 3:补签 4:每日任务 5:福利任务)
	Description     any         // 积分变动说明
	ExtJson         any         // 扩展字段
	CreatedAt       *gtime.Time // 创建时间
	UpdatedAt       *gtime.Time // 更新时间
	DeletedAt       *gtime.Time // 删除时间
}
