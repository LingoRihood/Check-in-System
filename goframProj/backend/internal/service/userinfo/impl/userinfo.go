package impl

import (
	"backend/internal/consts"
	"backend/internal/dao"
	"backend/internal/model"
	"backend/internal/model/entity"
	"backend/internal/service/userinfo"
	"backend/utility/injection"
	"context"
	"strconv"
	"time"

	"github.com/gogf/gf/crypto/gmd5"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sony/sonyflake/v2"
)

// 用户相关业务逻辑
// 定义一个结构体, 实现UserInfoService接口
type UserInfo struct {
	snowflack *sonyflake.Sonyflake
}

// encryptPassword 加密密码
func (u *UserInfo) encryptPassword(password string) string {
	return gmd5.MustEncryptString(password)
}

// func New() *UserInfo {
// 	return &UserInfo{}
// }

func New() userinfo.UserInfoService {
	return &UserInfo{
		snowflack: injection.MustInvoke[*sonyflake.Sonyflake](),
	}
}

// 建议：在包级别初始化一次（比如 init() 或 main 启动时）
// var sf *sonyflake.Sonyflake

const (
	defaultAvatar = "https://avatars.githubusercontent.com/u/51045272?v=4"
)

// func init() {
// 	st, err := time.Parse(time.DateOnly, "2025-11-01") // 用过去时间
// 	if err != nil {
// 		panic(err) // 这里只在启动阶段 panic 可以接受
// 	}
// 	sf, err = sonyflake.New(sonyflake.Settings{StartTime: st})
// 	if err != nil {
// 		panic(err)
// 	}
// }

// Create 创建用户
// func (u *UserInfo) Create(ctx context.Context, username, password, email string) error {
func (u *UserInfo) Create(ctx context.Context, input *model.CreateUserInput) (*model.CreateUserOutput, error) {
	// 1. 判断用户名是否已经存在（根据用户名查重）
	exist, err := dao.Userinfo.Ctx(ctx).
		Where(dao.Userinfo.Columns().Username, input.Username).
		Exist()
	if err != nil {
		g.Log().Errorf(ctx, "查询用户是否存在失败: %v", err)
		return nil, err
	}

	if exist {
		return nil, gerror.New("用户已存在")
	}

	// 2. 生成唯一 id
	userId, err := u.snowflack.NextID()
	if err != nil {
		g.Log().Errorf(ctx, "生成用户ID失败: %v", err)
		return nil, gerror.Wrap(err, "生成用户ID失败")
	}

	// 3. 创建用户
	// 创建用户，入库
	newUserInfo := entity.Userinfo{
		UserId:   uint64(userId), // 使用雪花算法生成唯一ID
		Username: input.Username,
		// Password: req.Password, // 是不是需要对用户输入的密码进行加密
		Password: u.encryptPassword(input.Password), // 对字符串 s 做 MD5, 返回 MD5 的十六进制字符串
		// 练习可用；生产换 bcrypt

		Email:  input.Email,
		Avatar: defaultAvatar, // 简化注册流程，一般使用默认头像，后续支持用户在个人中心上传头像
	}

	// id, err := dao.Userinfo.Ctx(ctx).InsertAndGetId(newUserInfo)

	_, err = dao.Userinfo.Ctx(ctx).Insert(newUserInfo)
	if err != nil {
		g.Log().Errorf(ctx, "创建用户失败：%v", err)
		return nil, gerror.Wrap(err, "创建用户失败")
	}
	// 4. 返回结果
	return &model.CreateUserOutput{
		UserId:   uint64(userId),
		Username: input.Username,
	}, nil
}

// Login 登录
func (u *UserInfo) Login(ctx context.Context, input *model.LoginInput) (*model.LoginOutput, error) {
	// 拿用户输入的用户名和密码，去数据库查询
	var user entity.Userinfo
	err := dao.Userinfo.Ctx(ctx).
		Where(dao.Userinfo.Columns().Username, input.Username).
		Where(dao.Userinfo.Columns().Password, u.encryptPassword(input.Password)).
		Scan(&user)

	if err != nil {
		g.Log().Errorf(ctx, "查询用户失败: %v", err)
		return nil, gerror.Wrapf(err, "查询用户失败")
	}

	// 生成 JWT Token
	tokenObj, err := genJwtByUserInfo(ctx, user.UserId, user.Username)
	if err != nil {
		g.Log().Errorf(ctx, "生成 JWT Token 失败: %v", err)
		return nil, gerror.Wrap(err, "生成 JWT Token 失败")
	}

	// 返回结果
	return &model.LoginOutput{
		AccessToken:  tokenObj.AccessToken,
		RefreshToken: tokenObj.RefreshToken,
	}, nil
}

// type JWTClaims struct {
// 	UserId   uint64 `json:"userId"`
// 	Username string `json:"username"`
// 	jwt.RegisteredClaims
// }

// genJwtByUserInfo 根据用户信息生成 JWT
func genJwtByUserInfo(ctx context.Context, userID uint64, username string) (*model.TokenOutput, error) {
	// 生成 Access Token
	claims := &model.JWTClaims{
		UserId:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Simon",
			Subject:   "check-in-system",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(consts.JWTTokenExpireDay * time.Hour)), // 设置过期时间为1天
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(consts.JWTTokenExpireSeconds * time.Second)), // 设置过期时间为20秒，【for the test sake】
		},
	}

	// 创建一个新的 JWT 对象
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// SignedString：把 Header + Payload 按 JWT 规则 base64url 编码后，使用 secret 生成签名，最后拼成标准 JWT 字符串
	// []byte(...)：HMAC 需要字节数组形式的 key
	signedAccessToken, err := accessToken.SignedString([]byte(consts.JWTAccessTokenSecret))
	if err != nil {
		g.Log().Errorf(ctx, "生成 JWT Token 失败: %v", err)
		return nil, err
	}

	// 生成 refresh Token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &model.JWTClaims{
		UserId:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Simon",
			Subject:   "check-in-system",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(consts.JWTRefreshExpireWeek * time.Hour)), // 设置过期时间为1周
		},
	})

	signedRefreshToken, err := refreshToken.SignedString([]byte(consts.JWTRefreshTokenSecret))
	if err != nil {
		g.Log().Errorf(ctx, "生成 JWT Token 失败: %v", err)
		return nil, err
	}

	return &model.TokenOutput{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
	}, nil
}

func (u *UserInfo) GetInfo(ctx context.Context, userId string) (*entity.Userinfo, error) {
	var user entity.Userinfo
	err := dao.Userinfo.Ctx(ctx).
		Where(dao.Userinfo.Columns().UserId, userId).
		Scan(&user)

	if err != nil {
		g.Log().Errorf(ctx, "查询用户失败: %v", err)
		return nil, gerror.Wrapf(err, "查询用户失败")
	}
	return &user, nil
}

// RefreshToken 刷新 Token
func (u *UserInfo) RefreshToken(ctx context.Context, refreshToken string) (res *model.TokenOutput, err error) {
	// 1. 解析 refreshToken，拿到 userid
	// 解析token获取用户id
	var claim model.JWTClaims

	// 解析 JWT：将 JWT 字符串（即 accessToken）分解为 Header、Payload、Signature
	// 验证签名：通过密钥 consts.JWTAccessTokenSecret 来验证 Signature 是否正确，确保 Token 没有被篡改。
	// 填充 claim：将解析出来的 Payload 内容（例如用户的 UserId、Username）填充到 claim 变量中
	// keyFunc：这个是一个回调函数，用来返回签名验证所需的密钥。jwt.ParseWithClaims 会使用它来验证 JWT 的 Signature 是否有效
	token, err := jwt.ParseWithClaims(refreshToken, &claim, func(token *jwt.Token) (any, error) {
		return []byte(consts.JWTRefreshTokenSecret), nil
	})

	if err != nil || !token.Valid {
		g.Log().Errorf(ctx, "refresh token: %v, err: %+v", token, err)
		return nil, gerror.New("refresh token 无效")
	}

	// 2. 根据 userid 获取用户信息
	userInfo, err := u.GetInfo(ctx, strconv.FormatUint(claim.UserId, 10))
	if err != nil {
		g.Log().Errorf(ctx, "根据 userid 获取用户信息失败: %v", err)
		return nil, gerror.Wrap(err, "根据 userid 获取用户信息失败")
	}

	// 3. 生成新的 accessToken 和 refreshToken
	// 4. 返回新的 token
	return genJwtByUserInfo(ctx, userInfo.UserId, userInfo.Username)
}
