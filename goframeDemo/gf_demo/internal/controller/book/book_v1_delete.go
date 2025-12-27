package book

import (
	"context"

	// dao 和 do：分别是数据访问对象（DAO）和数据库操作模型（DO）。dao.Book 用于执行与数据库交互的操作，do.Book 是数据库表的模型映射
	v1 "gf_demo/api/book/v1"
	"gf_demo/internal/dao"
)

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	// 删除图书的业务逻辑
	_, err = dao.Book.Ctx(ctx).WherePri(req.Id).Delete()
	return nil, err

	// return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
