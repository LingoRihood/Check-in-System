package consts

// CtxKey 上下文 key
type CtxKey string

const (
	JWTAccessTokenSecret         = "You're making my blood run beyond dimensions" // JWT 访问令牌密钥
	JWTRefreshTokenSecret        = "You're making my blood run out of emissions"  // JWT 刷新令牌密钥
	JWTTokenExpireSeconds        = 20                                             // JWT 令牌过期时间(秒)[为了测试而定义的常量]
	JWTTokenExpireDay            = 24                                             // JWT 令牌过期时间(小时)
	JWTRefreshExpireWeek         = 7 * 24                                         // JWT 刷新令牌过期时间(小时)
	CtxKeyUserID          CtxKey = "userId"                                       // 用户ID上下文 key
)
