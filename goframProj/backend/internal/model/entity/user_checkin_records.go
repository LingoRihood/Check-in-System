// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserCheckinRecords is the golang structure for table user_checkin_records.
type UserCheckinRecords struct {
	Id                uint64      `json:"id"                orm:"id"                  description:"记录ID"`               // 记录ID
	UserId            uint64      `json:"userId"            orm:"user_id"             description:"用户ID"`               // 用户ID
	CheckinDate       *gtime.Time `json:"checkinDate"       orm:"checkin_date"        description:"签到日期"`               // 签到日期
	CheckinType       int         `json:"checkinType"       orm:"checkin_type"        description:"签到类型: 1=正常签到, 2=补签"` // 签到类型: 1=正常签到, 2=补签
	PointsAwardedBase int         `json:"pointsAwardedBase" orm:"points_awarded_base" description:"获得积分"`               // 获得积分
	CreatedAt         *gtime.Time `json:"createdAt"         orm:"created_at"          description:"创建时间"`               // 创建时间
	UpdatedAt         *gtime.Time `json:"updatedAt"         orm:"updated_at"          description:"更新时间"`               // 更新时间
	DeletedAt         *gtime.Time `json:"deletedAt"         orm:"deleted_at"          description:"删除时间"`               // 删除时间
}
