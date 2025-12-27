package book

import (
	"context"

	// dao 和 do：分别是数据访问对象（DAO）和数据库操作模型（DO）。dao.Book 用于执行与数据库交互的操作，do.Book 是数据库表的模型映射
	v1 "gf_demo/api/book/v1"
	"gf_demo/internal/dao"
	"gf_demo/internal/model/do"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	// 更新书籍的业务逻辑
	_, err = dao.Book.Ctx(ctx).Data(do.Book{
		Title:  req.Title,
		Price:  req.Price,
		Status: req.Status,
	}).WherePri(req.Id).Update()

	return nil, err

	// return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
