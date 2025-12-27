// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Userinfo is the golang structure of table userinfo for DAO operations like Where/Data.
type Userinfo struct {
	g.Meta    `orm:"table:userinfo, do:true"`
	Id        any         // 主键ID
	UserId    any         // 用户ID
	Username  any         // 用户名
	Password  any         // 用户密码(md5加密)
	Email     any         // 用户邮箱
	Avatar    any         // 用户头像
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 删除时间
}
