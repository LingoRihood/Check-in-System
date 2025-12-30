package model

// 定义 积分 相关模型
type PointsTransactionInput struct {
	UserId uint64 // 用户ID
	Points int64  // 积分数量
	Desc   string // 描述
	Type   int    // 积分类型
}
