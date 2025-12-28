// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserPoints is the golang structure of table user_points for DAO operations like Where/Data.
type UserPoints struct {
	g.Meta      `orm:"table:user_points, do:true"`
	Id          any         // 主键ID
	UserId      any         // 用户ID
	Points      any         // 当前可用积分
	PointsTotal any         // 累计获得积分
	CreatedAt   *gtime.Time // 创建时间
	UpdatedAt   *gtime.Time // 更新时间
	DeletedAt   *gtime.Time // 删除时间
}
