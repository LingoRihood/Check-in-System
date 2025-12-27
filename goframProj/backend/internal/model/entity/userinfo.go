// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Userinfo is the golang structure for table userinfo.
type Userinfo struct {
	Id        uint64      `json:"id"        orm:"id"         description:"主键ID"`        // 主键ID
	UserId    uint64      `json:"userId"    orm:"user_id"    description:"用户ID"`        // 用户ID
	Username  string      `json:"username"  orm:"username"   description:"用户名"`         // 用户名
	Password  string      `json:"password"  orm:"password"   description:"用户密码(md5加密)"` // 用户密码(md5加密)
	Email     string      `json:"email"     orm:"email"      description:"用户邮箱"`        // 用户邮箱
	Avatar    string      `json:"avatar"    orm:"avatar"     description:"用户头像"`        // 用户头像
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`        // 创建时间
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`        // 更新时间
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"删除时间"`        // 删除时间
}
