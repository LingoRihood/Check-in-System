package userinfo

import (
	// "backend/internal/dao"
	"backend/internal/model"
	"backend/internal/model/entity"

	// "backend/internal/model/entity"
	"context"
	// "time"
	// "github.com/gogf/gf/crypto/gmd5"
	// "github.com/gogf/gf/v2/errors/gerror"
	// "github.com/gogf/gf/v2/frame/g"
	// "github.com/sony/sonyflake/v2"
)

// 把用户服务抽象成一个接口, 列出来所需要实现的方法
type UserInfoService interface {
	Create(ctx context.Context, input *model.CreateUserInput) (*model.CreateUserOutput, error)
	Login(ctx context.Context, input *model.LoginInput) (*model.LoginOutput, error)
	GetInfo(ctx context.Context, userId string) (*entity.Userinfo, error)
	RefreshToken(ctx context.Context, refreshToken string) (res *model.TokenOutput, err error)
}
