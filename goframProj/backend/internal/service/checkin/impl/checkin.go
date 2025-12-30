package impl

import (
	"backend/internal/dao"
	"backend/internal/model/entity"
	"backend/utility/injection"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/redis/go-redis/v9"
)

// 签到相关业务逻辑的具体实现

const (
	yearSignKeyFormat  = "user:checkins:daily:%d:%d" // user:checkins:daily:12131321421312:2025
	defaultDailyPoints = 1                           // 每日签到积分
)

type PointsTransactionType int

const (
	PointsTransactionTypeDaily       PointsTransactionType = iota + 1 // 每日签到 1
	PointsTransactionTypeConsecutive                                  // 连续签到 2
	PointsTransactionTypeRetro                                        // 补签 3
)

var PointsTransactionTypeMsgMap = map[PointsTransactionType]string{
	PointsTransactionTypeDaily:       "每日签到奖励",
	PointsTransactionTypeConsecutive: "连续签到奖励",
	PointsTransactionTypeRetro:       "补签消耗积分",
}

type Service struct {
	rc *redis.Client
}

func NewService() *Service {
	return &Service{
		rc: injection.MustInvoke[*redis.Client](), // 从注入器中获取 Redis 客户端实例
	}
}

// // Daily 每日签到
// func (s *Service) Daily(ctx context.Context, userId uint64) error {
// 	// 采用服务器时间进行每日签到，不依赖客户端传递的时间
// 	// 1. Redis 中使用 bitmap setbit 执行签到逻辑
// 	// 拿到当天是一年中的第几天，然后使用 setbit 记录这一天是否签到
// 	now := time.Now()
// 	year := now.Year()
// 	dayOfYearOffset := now.YearDay() - 1 // 因为 Redis bitmap 从 0 开始，所以要减一
// 	key := fmt.Sprintf(yearSignKeyFormat, userId, year)
// 	g.Log().Debugf(ctx, "key: %s dayOfYearOffset:%d", key, dayOfYearOffset)

// 	ret := s.rc.SetBit(ctx, key, int64(dayOfYearOffset), 1).Val()
// 	if ret == 1 {
// 		return errors.New("今日已签到")
// 	}

// 	// 2. 发放每日签到的积分
// 	// 用户积分汇总表 user_points 增加积分
// 	// 2.1 先查询(新用户可能没有记录)
// 	var userPoint entity.UserPoints
// 	if err := dao.UserPoints.Ctx(ctx).
// 		Where(dao.UserPoints.Columns().UserId, userId).
// 		Scan(&userPoint); err != nil && !errors.Is(err, sql.ErrNoRows) {
// 		g.Log().Errorf(ctx, "查询用户积分汇总表失败: %v", err)
// 		return err
// 	}

// 	// 如果查不到，则插入一条记录
// 	if userPoint.Id == 0 {
// 		userPoint = entity.UserPoints{UserId: userId} // 创建新对象
// 	}

// 	userPoint.Points = userPoint.Points + defaultDailyPoints           // 增加每日签到积分
// 	userPoint.PointsTotal = userPoint.PointsTotal + defaultDailyPoints // 累计积分

// 	// 2.2 事务更新 用户积分汇总表 和 用户积分明细表
// 	// 为什么要事务？
// 	// 因为要保证：流水插入成功, 汇总更新成功
// 	// 这两件事要么都成功，要么都失败，不然会出现“余额变了但没流水”或“有流水但余额没变”。
// 	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
// 		// 用户积分明细表 user_points_transactions 增加记录
// 		newRecord := entity.UserPointsTransactions{
// 			UserId:          userId,
// 			PointsChange:    defaultDailyPoints,
// 			CurrentBalance:  userPoint.Points,
// 			TransactionType: int(PointsTransactionTypeDaily),
// 			Description:     PointsTransactionTypeMsgMap[PointsTransactionTypeDaily],
// 			CreatedAt:       gtime.NewFromTime(time.Now()),
// 			UpdatedAt:       gtime.NewFromTime(time.Now()),
// 		}

// 		// return nil => 提交 commit
// 		// return err => 回滚 rollback
// 		// tx.Model(...)：明确使用事务 tx 执行 SQL
// 		if _, err := tx.Model(&entity.UserPointsTransactions{}).Insert(&newRecord); err != nil {
// 			g.Log().Errorf(ctx, "插入用户积分明细表失败: %v", err)
// 			return err
// 		}
// 		if _, err := tx.Model(&entity.UserPoints{}).
// 			Where(dao.UserPoints.Columns().UserId, userId).
// 			Save(&userPoint); err != nil {
// 			g.Log().Errorf(ctx, "更新用户积分汇总表失败: %v", err)
// 			return err
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		g.Log().Errorf(ctx, "事务处理失败: %v", err)
// 		return err
// 	}

// 	// 3. 发送连续签到的奖励积分

// 	return nil
// }

// Daily 每日签到（先 DB 成功，再 setbit）
// ✅ 先 DB 加分成功
// ✅ 再 Redis 标记已签到
// 这样就不会出现“签了但没加分”
// 关键点：
// 1) 先 GetBit 判断是否已签到（只读）
// 2) 事务里：写明细 + 更新/插入汇总（先 DB）
// 3) DB 成功后：SetBit 标记已签到
// 4) 用 Redis SetNX 做锁，防止并发重复加分
func (s *Service) Daily(ctx context.Context, userId uint64) error {
	// 采用服务器时间进行每日签到，不依赖客户端传递的时间
	// 1. Redis 中使用 bitmap setbit 执行签到逻辑
	// 拿到当天是一年中的第几天，然后使用 setbit 记录这一天是否签到
	now := time.Now()
	year := now.Year()
	dayOfYearOffset := now.YearDay() - 1 // 因为 Redis bitmap 从 0 开始，所以要减1
	signKey := fmt.Sprintf(yearSignKeyFormat, userId, year)

	// 0) 防并发：同一用户同一天只允许一个请求进来（10s 超时防死锁）
	lockKey := fmt.Sprintf("lock:checkins:daily:%d:%d:%d", userId, year, dayOfYearOffset)
	locked, err := s.rc.SetNX(ctx, lockKey, 1, 10*time.Second).Result()
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("签到处理中，请稍后重试")
	}
	defer s.rc.Del(ctx, lockKey)

	// 1) 只读判断：是否已签到
	bit, err := s.rc.GetBit(ctx, signKey, int64(dayOfYearOffset)).Result()
	if err != nil {
		return err
	}
	if bit == 1 {
		return errors.New("今日已签到")
	}

	// 2) 先 DB：事务里写明细 + 更新/插入汇总
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 2.1 查询积分汇总（新用户可能没有记录）
		var userPoint entity.UserPoints
		if err := tx.Model(&entity.UserPoints{}).
			Where(dao.UserPoints.Columns().UserId, userId).
			Scan(&userPoint); err != nil && !errors.Is(err, sql.ErrNoRows) {
			g.Log().Errorf(ctx, "查询用户积分汇总表失败: %v", err)
			return err
		}

		isNew := userPoint.Id == 0
		if isNew {
			userPoint = entity.UserPoints{
				UserId: userId,
			}
		}

		// 2.2 计算加分后的值
		userPoint.Points += defaultDailyPoints
		userPoint.PointsTotal += defaultDailyPoints

		// 2.3 插入积分明细
		newRecord := entity.UserPointsTransactions{
			UserId:          userId,
			PointsChange:    defaultDailyPoints,
			CurrentBalance:  userPoint.Points,
			TransactionType: int(PointsTransactionTypeDaily), // ✅ entity 字段是 int，显式转换
			Description:     PointsTransactionTypeMsgMap[PointsTransactionTypeDaily],
			CreatedAt:       gtime.NewFromTime(now),
			UpdatedAt:       gtime.NewFromTime(now),
		}

		if _, err := tx.Model(&entity.UserPointsTransactions{}).Insert(&newRecord); err != nil {
			g.Log().Errorf(ctx, "插入用户积分明细表失败: %v", err)
			return err
		}

		// 2.4 更新/插入积分汇总
		if isNew {
			userPoint.CreatedAt = gtime.NewFromTime(now)
			userPoint.UpdatedAt = gtime.NewFromTime(now)

			if _, err := tx.Model(&entity.UserPoints{}).Insert(&userPoint); err != nil {
				g.Log().Errorf(ctx, "插入用户积分汇总表失败: %v", err)
				return err
			}
		} else {
			if _, err := tx.Model(&entity.UserPoints{}).
				Where(dao.UserPoints.Columns().UserId, userId).
				Data(g.Map{
					dao.UserPoints.Columns().Points:      userPoint.Points,
					dao.UserPoints.Columns().PointsTotal: userPoint.PointsTotal,
					dao.UserPoints.Columns().UpdatedAt:   gtime.NewFromTime(now),
				}).
				Update(); err != nil {
				g.Log().Errorf(ctx, "更新用户积分汇总表失败: %v", err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		g.Log().Errorf(ctx, "事务处理失败: %v", err)
		return err
	}

	// 3) 再 Redis：DB 成功后再 setbit（避免“签到了但没加分”）
	old, err := s.rc.SetBit(ctx, signKey, int64(dayOfYearOffset), 1).Result()
	if err != nil {
		return err
	}
	// 理论上有锁不会发生，兜底
	if old == 1 {
		return errors.New("今日已签到")
	}

	return nil
}
