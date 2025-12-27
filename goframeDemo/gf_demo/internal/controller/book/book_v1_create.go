package book

import (
	"context"

	_ "github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	_ "github.com/gogf/gf/v2/frame/g"

	// dao 和 do：分别是数据访问对象（DAO）和数据库操作模型（DO）。dao.Book 用于执行与数据库交互的操作，do.Book 是数据库表的模型映射
	v1 "gf_demo/api/book/v1"
	"gf_demo/internal/dao"
	"gf_demo/internal/model/do"
)

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	// 1. 解析请求参数并对请求参数进行校验
	// 2. 实现创建图书的业务逻辑
	// dao.Book.Ctx(ctx)：调用 dao 中 Book 对象的上下文方法，表示在当前上下文中执行数据库操作。
	// .Data(do.Book{...})：这里的 do.Book 是将传入的 req.Title 和 req.Price 以及默认的 Status（StatusAvailable）放入一个 do.Book 结构体中。do.Book 是用于与数据库表映射的模型，包含了书籍的字段。
	// .InsertAndGetId()：插入新的书籍记录到数据库，并返回插入数据的 ID（lastInsertId）。如果插入失败，err 会包含错误信息
	lastInsertId, err := dao.Book.Ctx(ctx).Data(do.Book{
		Title:  req.Title,
		Price:  req.Price,
		Status: v1.StatusAvailable, // 默认可用状态
	}).InsertAndGetId()

	// 等价于 INSERT INTO book (title, price, status) VALUES (?, ?, ?);
	// g.DB(): 获取默认的数据库对象（连接池/操作入口）
	// .Model("book"): 指定要操作的表是 book
	// 把 Go 的 context 传进来。作用常见有：超时控制（ctx 超时就取消 DB 操作）、链路追踪/日志关联、请求结束自动取消等
	// .Data(do.Book{...}): 指定“要插入的数据”
	// .InsertAndGetId(): 执行插入，并返回 插入记录的主键ID（通常是自增 ID）
	// lastInsertId, err := g.DB().Model("book").Ctx(ctx).Data(do.Book{
	// 	Title:  req.Title,
	// 	Price:  req.Price,
	// 	Status: v1.StatusAvailable, // 默认可用状态
	// }).InsertAndGetId()

	// lastInsertId, err := g.DB().Model("book").Ctx(ctx).Data(g.Map{
	// 	"title":  req.Title,
	// 	"price":  req.Price,
	// 	"status": v1.StatusAvailable, // 默认可用状态
	// }).InsertAndGetId()

	if err != nil {
		return nil, gerror.Wrap(err, "failed to create book")
	}

	// 3. 返回创建结果
	return &v1.CreateRes{
		Id: lastInsertId, // 假设创建的图书ID为1
	}, nil

	// 创建了一个自定义错误，通常用于未实现的功能或者接口。
	// return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
