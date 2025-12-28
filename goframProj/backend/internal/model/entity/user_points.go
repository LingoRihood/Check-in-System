// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserPoints is the golang structure for table user_points.
type UserPoints struct {
	Id          uint64      `json:"id"          orm:"id"           description:"主键ID"`   // 主键ID
	UserId      uint64      `json:"userId"      orm:"user_id"      description:"用户ID"`   // 用户ID
	Points      int64       `json:"points"      orm:"points"       description:"当前可用积分"` // 当前可用积分
	PointsTotal int64       `json:"pointsTotal" orm:"points_total" description:"累计获得积分"` // 累计获得积分
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:"创建时间"`   // 创建时间
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:"更新时间"`   // 更新时间
	DeletedAt   *gtime.Time `json:"deletedAt"   orm:"deleted_at"   description:"删除时间"`   // 删除时间
}
