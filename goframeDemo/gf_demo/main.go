package main

import (
	_ "gf_demo/internal/packed"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2" // 导入MySQL驱动

	_ "github.com/gogf/gf/contrib/nosql/redis/v2" // 导入Redis驱动

	"github.com/gogf/gf/v2/os/gctx"

	"gf_demo/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
