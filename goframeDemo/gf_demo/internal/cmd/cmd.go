package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gcmd"

	"gf_demo/internal/controller/book"
	"gf_demo/internal/controller/hello"
)

// func MiddlewareAuth(r *ghttp.Request) {
// 	// 在请求处理之前进行鉴权
// 	// 根据请求是否携带某些token，来判断是否是有效请求
// 	token := r.Get("token")
// 	if token.String() == "123" {
// 		// “放行”，继续执行后面的流程（下一个中间件 / 业务 handler / controller）。
// 		r.Middleware.Next()
// 		return
// 	}

// 	// 鉴权失败，返回403状态码
// 	r.Response.WriteStatus(403)
// }

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			// s.Use() // 注册全局中间件
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse) // 分组注册中间件
				group.Bind(
					hello.NewV1(),

					// 注册book相关路由
					// 请求方法和路径是根据 请求结构体里的 g.Meta 标签自动生成的
					book.NewV1(),
				)
			})

			// s.Use(MiddlewareAuth)
			// s.BindHandler("GET:/index", func(r *ghttp.Request) {
			// 	r.Response.WriteJsonExit(map[string]string{
			// 		"name": "index",
			// 	})
			// })

			// 给服务器绑定一个路由：当有人用 GET 方法访问 /login 时，就返回一段固定的 JSON，并立刻结束请求
			// s.BindHandler("GET:/login", func(r *ghttp.Request) {
			// 	// 把你传进去的数据转换成 JSON, 写到 HTTP 响应里,立刻终止后续处理（不再继续执行后面的 handler / 中间件链）
			// 	r.Response.WriteJsonExit(map[string]string{
			// 		"name":  "Simon",
			// 		"value": "gf_demo",
			// 	})
			// })

			g.DB().GetCache().SetAdapter(gcache.NewAdapterRedis(g.Redis()))

			s.Run()
			return nil
		},
	}
)
