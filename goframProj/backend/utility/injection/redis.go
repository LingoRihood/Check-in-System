package injection

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
)

// injectRedis 注入 Redis client
func injectRedis(ctx context.Context, injector *do.Injector) {
	do.Provide(injector, func(i *do.Injector) (*redis.Client, error) {
		// 结构体字段必须大写：反射才能赋值
		// 配置键可以小写：GF Scan 会自动映射到大写字段名
		type RedisConfig struct {
			Address  string
			Password string
		}
		var (
			err    error
			config *RedisConfig
		)
		err = g.Cfg().MustGet(ctx, "redis.checkin").Scan(&config)
		if err != nil {
			return nil, err
		}
		if config == nil {
			return nil, gerror.New("redis config not found")
		}
		// g.Log().Debugf(ctx, "Redis Config: %+v", config)
		g.Log().Debugf(ctx, "Redis Config: addr=%s", config.Address)
		svc := redis.NewClient(&redis.Options{
			Addr:     config.Address,
			Password: config.Password,
		})

		// 注册“优雅退出”关闭 Redis
		SetupShutdownHelper(injector, svc, func(svc *redis.Client) error {
			return svc.Close()
		})
		return svc, nil
	})
}
