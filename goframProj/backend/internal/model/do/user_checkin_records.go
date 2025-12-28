// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserCheckinRecords is the golang structure of table user_checkin_records for DAO operations like Where/Data.
type UserCheckinRecords struct {
	g.Meta            `orm:"table:user_checkin_records, do:true"`
	Id                any         // 记录ID
	UserId            any         // 用户ID
	CheckinDate       *gtime.Time // 签到日期
	CheckinType       any         // 签到类型: 1=正常签到, 2=补签
	PointsAwardedBase any         // 获得积分
	CreatedAt         *gtime.Time // 创建时间
	UpdatedAt         *gtime.Time // 更新时间
	DeletedAt         *gtime.Time // 删除时间
}
