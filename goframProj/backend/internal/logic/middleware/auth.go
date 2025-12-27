package middleware

import (
	"backend/internal/consts"
	"backend/internal/model"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v5"
)

// Auth 认证中间件
func Auth(r *ghttp.Request) {
	// 从请求中获取用户id(根据Access token获取用户id)
	// 从请求头中获取 jwt access token
	// r := g.RequestFromCtx(ctx) // 从 ctx 中获取请求对象

	// 获取请求上下文
	ctx := r.GetCtx()

	// 从请求头中获取 Authorization 字段的值
	authorizationValue := r.GetHeader("Authorization")

	// HTTP 请求头里通常会带：Authorization: Bearer <access_token>
	// 检查 Authorization 头值是否以 Bearer 开头（Bearer 是 OAuth2.0 中用于表示携带 JWT 的标准方式）。
	if len(authorizationValue) == 0 || !strings.HasPrefix(authorizationValue, "Bearer ") {
		// r.Response.WriteStatusExit(http.StatusForbidden, "缺少token")

		// 401 未认证 403未授权
		r.Response.WriteStatusExit(http.StatusUnauthorized, "缺少token")
	}

	accessToken := strings.TrimPrefix(authorizationValue, "Bearer ")

	// 解析token获取用户id
	var claim model.JWTClaims

	// 解析 JWT：将 JWT 字符串（即 accessToken）分解为 Header、Payload、Signature
	// 验证签名：通过密钥 consts.JWTAccessTokenSecret 来验证 Signature 是否正确，确保 Token 没有被篡改。
	// 填充 claim：将解析出来的 Payload 内容（例如用户的 UserId、Username）填充到 claim 变量中
	// keyFunc：这个是一个回调函数，用来返回签名验证所需的密钥。jwt.ParseWithClaims 会使用它来验证 JWT 的 Signature 是否有效
	token, err := jwt.ParseWithClaims(accessToken, &claim, func(token *jwt.Token) (any, error) {
		return []byte(consts.JWTAccessTokenSecret), nil
	})

	if err != nil || !token.Valid {
		// 401 未认证 403未授权
		r.Response.WriteStatusExit(http.StatusUnauthorized, "无效的token")
	}

	g.Log().Debugf(ctx, "claim: %v", claim)

	// 向请求的上下文中写入用户 id
	r.SetCtxVar(consts.CtxKeyUserID, claim.UserId) // 使用自定义 Key 类型，防止被其他中间件覆盖
	// r.SetCtxVar("userId", "xxx")

	r.Middleware.Next() // 继续执行后续中间件
}
