package injection

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/samber/do"
	"github.com/sony/sonyflake/v2"
)

func injectSnowflake(ctx context.Context, injector *do.Injector) {
	// 往 do.Injector 容器里注册一个依赖：*sonyflake.Sonyflake
	// 在创建成功后，把它注册到“关闭钩子”里（SetupShutdownHelper），程序退出时可以做清理（虽然 sonyflake 一般不需要清理）
	do.Provide(injector, func(i *do.Injector) (*sonyflake.Sonyflake, error) {
		// 完成 *sonyflake.Sonyflake 的初始化
		// 1. 读取配置文件中的起始时间
		type SnowflakeCfg struct {
			StartTime string `yaml:"start_time"`
		}

		var (
			err error
			cfg *SnowflakeCfg
		)

		// .Cfg()：拿到配置对象
		// MustGet(ctx, "snowflake")：取配置节点 snowflake
		// .Scan(&cfg)：把节点内容映射到 cfg（结构体指针）
		err = g.Cfg().MustGet(ctx, "snowflake").Scan(&cfg)
		if err != nil {
			g.Log().Errorf(ctx, "manifest/config/config.yaml中必须指定snowflake的起始时间！: %v", err)
			return nil, err
		}

		if cfg == nil {
			return nil, fmt.Errorf("manifest/config/config.yaml中必须指定snowflake的起始时间！")
		}

		st, err := time.Parse(time.DateOnly, cfg.StartTime)
		if err != nil {
			return nil, fmt.Errorf("snowflake.start_time 格式必须是 YYYY-MM-DD: %w", err)
		}

		// 创建 sonyflake 实例
		settings := sonyflake.Settings{
			StartTime: st,
		}

		sonyFlake, err := sonyflake.New(settings)
		if err != nil {
			g.Log().Errorf(ctx, "初始化sonyflake失败: %v", err)
			return nil, err
		}

		// 注册 sonyflake 和一个关闭函数
		SetupShutdownHelper(injector, sonyFlake, func(svc *sonyflake.Sonyflake) error {
			return nil
		})

		return sonyFlake, nil
	})
}
