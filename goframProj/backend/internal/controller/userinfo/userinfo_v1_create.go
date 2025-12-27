package userinfo

import (
	"context"
	_ "time"

	"backend/internal/model"

	_ "github.com/gogf/gf/crypto/gmd5"
	_ "github.com/gogf/gf/v2/errors/gcode"
	_ "github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "backend/api/userinfo/v1"
	_ "backend/internal/dao"
	_ "backend/internal/model/entity"

	_ "github.com/sony/sonyflake/v2"
)

// const (
// 	defaultAvatar = "https://avatars.githubusercontent.com/u/51045272?v=4"
// )

// // 建议：在包级别初始化一次（比如 init() 或 main 启动时）
// var sf *sonyflake.Sonyflake

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

// 用户注册流程
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	// 参数校验（框架已经做完了）

	// st, _ := time.Parse(time.DateOnly, "2025-11-01")
	// settings := sonyflake.Settings{
	// 	StartTime: st,
	// }

	// sonyFlake, err := sonyflake.New(settings)
	// if err != nil {
	// 	panic(err)
	// }

	// userID, err := sonyFlake.NextID() // id
	// fmt.Printf("id:%v err:%v\n", userID, err)

	// userID, err := sf.NextID()
	// if err != nil {
	// 	g.Log().Errorf(ctx, "生成用户ID失败: %v", err)
	// 	return nil, gerror.Wrap(err, "生成用户ID失败")
	// }

	// 拿一条记录/返回给业务/插入一整条数据 → 用 entity.Userinfo
	// 写查询条件、更新字段（部分更新） → 用 do.Userinfo
	// newUserInfo := entity.Userinfo{
	// 	UserId:   uint64(userID), // 使用雪花算法生成唯一ID
	// 	Username: req.Username,
	// 	// Password: req.Password, // 是不是需要对用户输入的密码进行加密
	// 	Password: gmd5.MustEncryptString(req.Password), // 对字符串 s 做 MD5, 返回 MD5 的十六进制字符串
	// 	// 练习可用；生产换 bcrypt

	// 	Email:  req.Email,
	// 	Avatar: defaultAvatar, // 简化注册流程，一般使用默认头像，后续支持用户在个人中心上传头像
	// }

	// // id, err := dao.Userinfo.Ctx(ctx).InsertAndGetId(newUserInfo)

	// _, err = dao.Userinfo.Ctx(ctx).Insert(newUserInfo)
	// if err != nil {
	// 	g.Log().Errorf(ctx, "创建用户失败：%v", err)
	// 	return nil, gerror.Wrap(err, "创建用户失败")
	// }

	// 返回结果
	// return &v1.CreateRes{
	// 	UserId:   uint64(userID),
	// 	Username: newUserInfo.Username,
	// }, nil

	// 参数校验（框架已经做完了）
	input := model.CreateUserInput{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}

	// 调用service层的逻辑
	output, err := c.svc.Create(ctx, &input)
	if err != nil {
		g.Log().Errorf(ctx, "Create user failed: %v", err)
		return nil, err
	}

	// 返回结果
	return &v1.CreateRes{
		UserId:   output.UserId,
		Username: output.Username,
	}, nil
}
