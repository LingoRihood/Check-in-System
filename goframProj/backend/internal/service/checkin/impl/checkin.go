package impl

// 签到相关业务逻辑的具体实现

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// Daily 每日签到
func (Service) Daily() {
	// 采用服务器时间进行每日签到，不依赖客户端传递的时间
	// 1. Redis 中使用 bitmap setbit 执行签到逻辑
	// 2. 发放每日签到的积分
	// 3. 发送连续签到的奖励积分

}
