package book

import (
	"context"
	"fmt"
	"time"

	// dao 和 do：分别是数据访问对象（DAO）和数据库操作模型（DO）。dao.Book 用于执行与数据库交互的操作，do.Book 是数据库表的模型映射
	v1 "gf_demo/api/book/v1"
	"gf_demo/internal/dao"
	"gf_demo/internal/model/do"
	"gf_demo/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	// 获取图书列表的业务逻辑
	// 结构体的字段会使用其 零值：
	// List 字段的零值是 nil，因为 List 是一个切片类型，切片的零值是 nil
	// res = &v1.GetListRes{}

	// 内存缓存
	cache := gcache.New()

	// 把 key1 存起来，10 分钟后自动过期删除
	cache.Set(ctx, "key1", "value1", time.Minute*10)

	// 返回的是 gvar.Var类型
	// if v, err := cache.Get(ctx, "key1"); err != nil {
	// 	return nil, err
	// } else {
	// 	fmt.Println(v.String())
	// }

	v, err := cache.Get(ctx, "key1")
	if err != nil {
		return nil, err
	}
	if v == nil || v.IsNil() {
		fmt.Println("缓存不存在或已过期")
		return nil, nil
	}
	fmt.Println(v.String())

	// 使用gcache包提供的全局cache对象
	// gcache.Set(ctx, "key2", "value2", time.Minute * 10)

	// 基于redis的缓存
	redisCache := gcache.New()
	// 拿到redis client: gredis.Redis
	redisCache.SetAdapter(gcache.NewAdapterRedis(g.Redis())) // 设置Redis适配器
	redisCache.Set(ctx, "key2", "value2", time.Minute*10)

	v, err = redisCache.Get(ctx, "key2")
	if err != nil {
		return nil, err
	}
	if v == nil || v.IsNil() {
		fmt.Println("Redis缓存不存在或已过期")
		return nil, nil
	}
	fmt.Println(v.String())

	res = &v1.GetListRes{
		List: make([]*entity.Book, 0), // 初始化为长度为 0 的切片
	}

	// Where(...) 负责筛选条件，Scan(...) 负责把数据库结果“抄写”进你的变量里。
	// 这是“按条件过滤”：只查 status = req.Status 的记录
	// 等价于 SQL 的：SELECT * FROM book WHERE status = ?
	// （? 就是 req.Status）
	// 把查询出来的“多行记录”填充到 res.List 这个切片里
	// err = dao.Book.Ctx(ctx).
	// 	Where(do.Book{
	// 		Status: req.Status,
	// 	}).Scan(&res.List)

	// 默认使用内存缓存
	// do.Book 是数据库表字段的映射模型（Data Object）。
	// 这里等价于 SQL：WHERE status = ?（? 的值就是 req.Status）
	// Scan 会把查询的多行结果扫描到 res.List 里（切片里每个元素是 *entity.Book）
	// 去 book 表里查“状态 = req.Status”的书，把查到的结果塞进 res.List；并且把这次查询结果缓存 10 分钟，10 分钟内，如果有人再发起同样的查询，就直接从缓存拿结果，不再查数据库
	// 这次查出来的结果，帮我存到缓存里，存 10 分钟。缓存这件事起个名字叫 book_list
	err = dao.Book.Ctx(ctx).Cache(gdb.CacheOption{
		Name:     "book_list",
		Duration: time.Minute * 10,
	}).Where(do.Book{
		// “只要 status 等于 req.Status 的那些书。”
		// 等价 SQL：SELECT * FROM book WHERE status = ?
		Status: req.Status,
	}).Scan(&res.List)
	return res, err
}
