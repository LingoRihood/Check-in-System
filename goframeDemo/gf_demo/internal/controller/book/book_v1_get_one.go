package book

import (
	"context"

	// dao 和 do：分别是数据访问对象（DAO）和数据库操作模型（DO）。dao.Book 用于执行与数据库交互的操作，do.Book 是数据库表的模型映射
	v1 "gf_demo/api/book/v1"
	"gf_demo/internal/dao"
)

func (c *ControllerV1) GetOne(ctx context.Context, req *v1.GetOneReq) (res *v1.GetOneRes, err error) {
	// 查询单个图书的业务逻辑
	res = &v1.GetOneRes{} // ✅ 关键：先分配内存

	// Scan 的作用是：把数据库查到的一行记录，填充到你给它的变量里
	err = dao.Book.Ctx(ctx).
		WherePri(req.Id).
		Scan(&res.Book) // 查询到的数据赋值给res.Book变量

	return res, err
	// return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
