// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserinfoDao is the data access object for the table userinfo.
type UserinfoDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  UserinfoColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// UserinfoColumns defines and stores column names for the table userinfo.
type UserinfoColumns struct {
	Id        string // 主键ID
	UserId    string // 用户ID
	Username  string // 用户名
	Password  string // 用户密码(md5加密)
	Email     string // 用户邮箱
	Avatar    string // 用户头像
	CreatedAt string // 创建时间
	UpdatedAt string // 更新时间
	DeletedAt string // 删除时间
}

// userinfoColumns holds the columns for the table userinfo.
var userinfoColumns = UserinfoColumns{
	Id:        "id",
	UserId:    "user_id",
	Username:  "username",
	Password:  "password",
	Email:     "email",
	Avatar:    "avatar",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewUserinfoDao creates and returns a new DAO object for table data access.
func NewUserinfoDao(handlers ...gdb.ModelHandler) *UserinfoDao {
	return &UserinfoDao{
		group:    "default",
		table:    "userinfo",
		columns:  userinfoColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserinfoDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserinfoDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserinfoDao) Columns() UserinfoColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserinfoDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserinfoDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *UserinfoDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
