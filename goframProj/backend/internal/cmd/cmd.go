package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"backend/internal/controller/checkin"
	"backend/internal/controller/hello"
	"backend/internal/controller/userinfo"
	"backend/internal/logic/middleware"
	"backend/utility/injection"
)

var (
	// 运行 main 命令时，程序会进入 gcmd.Command 的 Func 函数，这是你定义的 main 命令的核心部分。
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 在 Func 函数内部，首先初始化了一个 GoFrame 服务器
			s := g.Server()

			// 服务注入
			injection.SetupDefaultInjector(ctx)
			defer injection.ShutdownDefaultInjector()

			// 定义了一个路由组，所有的路由都会以 /api/v1 为前缀。
			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				// 注册通用响应中间件和CORS跨域中间件
				// ghttp.MiddlewareHandlerResponse 是 GoFrame 的默认响应中间件，负责处理 HTTP 响应的通用逻辑。
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.CORS)

				// 不需要登录也能访问的接口
				group.POST("/auth/login", userinfo.NewV1(), "Login")          // 登录
				group.POST("/users", userinfo.NewV1(), "Create")              // 创建用户
				group.POST("/auth/refresh", userinfo.NewV1(), "RefreshToken") // 刷新token

				// 需要登录才能访问的接口
				group.Middleware(middleware.Auth)
				group.GET("/users/me", userinfo.NewV1(), "Me") // 我的信息

				group.Bind(
					hello.NewV1(),
					// userinfo.NewV1(), // 用户模块相关接口
					checkin.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}
)
