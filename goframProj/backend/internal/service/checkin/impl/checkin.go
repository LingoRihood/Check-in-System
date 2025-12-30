package impl

import (
	"backend/internal/dao"
	"backend/internal/model"
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
	yearSignKeyFormat         = "user:checkins:daily:%d:%d"     // user:checkins:daily:12131321421312:2025
	monthRetroKeyFormat       = "user:checkins:retro:%d:%d%02d" // user:checkins:retro:12131321421231202501
	defaultDailyPoints  int64 = 1                               // 每日签到积分（注意：用 int64，和积分字段统一）
)

type PointsTransactionType int

const (
	PointsTransactionTypeDaily       PointsTransactionType = iota + 1 // 每日签到 1
	PointsTransactionTypeConsecutive                                  // 连续签到 2
	PointsTransactionTypeRetro                                        // 补签 3
)

type ConsecutiveBonusType int32

const (
	// 连续签到奖励规则
	consecutiveBonus3  ConsecutiveBonusType = 1 // "连续签到3天奖励"
	consecutiveBonus7  ConsecutiveBonusType = 2 // "连续签到7天奖励"
	consecutiveBonus15 ConsecutiveBonusType = 3 // "连续签到15天奖励"
	consecutiveBonus30 ConsecutiveBonusType = 4 // "月度满签奖励"
)

var consecutiveBonusNames = map[ConsecutiveBonusType]string{
	consecutiveBonus3:  "连续签到3天奖励",
	consecutiveBonus7:  "连续签到7天奖励",
	consecutiveBonus15: "连续签到15天奖励",
	consecutiveBonus30: "月度满签奖励",
}

var PointsTransactionTypeMsgMap = map[PointsTransactionType]string{
	PointsTransactionTypeDaily:       "每日签到奖励",
	PointsTransactionTypeConsecutive: "连续签到奖励",
	PointsTransactionTypeRetro:       "补签消耗积分",
}

// consecutiveBonusRule 连续签到奖励规则
type ConsecutiveBonusRule struct {
	TriggerDays int                  // 触发连续签到奖励的天数
	Points      int64                // 连续签到奖励的积分
	BonusType   ConsecutiveBonusType // 连续签到奖励类型
}

var consecutiveBonusRules = []ConsecutiveBonusRule{
	{TriggerDays: 3, Points: 5, BonusType: consecutiveBonus3},
	{TriggerDays: 7, Points: 10, BonusType: consecutiveBonus7},
	{TriggerDays: 15, Points: 20, BonusType: consecutiveBonus15},
	{TriggerDays: 30, Points: 100, BonusType: consecutiveBonus30},
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
	// err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
	// 	// 2.1 查询积分汇总（新用户可能没有记录）
	// 	var userPoint entity.UserPoints
	// 	if err := tx.Model(&entity.UserPoints{}).
	// 		Where(dao.UserPoints.Columns().UserId, userId).
	// 		Scan(&userPoint); err != nil && !errors.Is(err, sql.ErrNoRows) {
	// 		g.Log().Errorf(ctx, "查询用户积分汇总表失败: %v", err)
	// 		return err
	// 	}

	// 	isNew := userPoint.Id == 0
	// 	if isNew {
	// 		userPoint = entity.UserPoints{
	// 			UserId: userId,
	// 		}
	// 	}

	// 	// 2.2 计算加分后的值
	// 	userPoint.Points += defaultDailyPoints
	// 	userPoint.PointsTotal += defaultDailyPoints

	// 	// 2.3 插入积分明细
	// 	newRecord := entity.UserPointsTransactions{
	// 		UserId:          userId,
	// 		PointsChange:    defaultDailyPoints,
	// 		CurrentBalance:  userPoint.Points,
	// 		TransactionType: int(PointsTransactionTypeDaily), // ✅ entity 字段是 int，显式转换
	// 		Description:     PointsTransactionTypeMsgMap[PointsTransactionTypeDaily],
	// 		CreatedAt:       gtime.NewFromTime(now),
	// 		UpdatedAt:       gtime.NewFromTime(now),
	// 	}

	// 	if _, err := tx.Model(&entity.UserPointsTransactions{}).Insert(&newRecord); err != nil {
	// 		g.Log().Errorf(ctx, "插入用户积分明细表失败: %v", err)
	// 		return err
	// 	}

	// 	// 2.4 更新/插入积分汇总
	// 	if isNew {
	// 		userPoint.CreatedAt = gtime.NewFromTime(now)
	// 		userPoint.UpdatedAt = gtime.NewFromTime(now)

	// 		if _, err := tx.Model(&entity.UserPoints{}).Insert(&userPoint); err != nil {
	// 			g.Log().Errorf(ctx, "插入用户积分汇总表失败: %v", err)
	// 			return err
	// 		}
	// 	} else {
	// 		if _, err := tx.Model(&entity.UserPoints{}).
	// 			Where(dao.UserPoints.Columns().UserId, userId).
	// 			Data(g.Map{
	// 				dao.UserPoints.Columns().Points:      userPoint.Points,
	// 				dao.UserPoints.Columns().PointsTotal: userPoint.PointsTotal,
	// 				dao.UserPoints.Columns().UpdatedAt:   gtime.NewFromTime(now),
	// 			}).
	// 			Update(); err != nil {
	// 			g.Log().Errorf(ctx, "更新用户积分汇总表失败: %v", err)
	// 			return err
	// 		}
	// 	}

	// 	return nil
	// })

	// 2. 发放每日签到的积分
	// 先写 DB 加积分
	if err := s.AddPoints(ctx, &model.PointsTransactionInput{
		UserId: userId,
		Points: defaultDailyPoints,
		Desc:   PointsTransactionTypeMsgMap[PointsTransactionTypeDaily],
		Type:   int(PointsTransactionTypeDaily),
	}); err != nil {
		g.Log().Errorf(ctx, "AddPoints 事务处理失败: %v", err)
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

	// 3. 发送连续签到的奖励积分
	return s.updateConsecutiveBonus(ctx, userId, year, int(now.Month()))
}

// updateConsecutiveBonus 更新连续签到奖励积分
func (s *Service) updateConsecutiveBonus(ctx context.Context, userId uint64, year, month int) error {
	// 1. 获取当前本月连续签到天数
	maxConsecutive, err := s.CalcMonthConsecutiveDays(ctx, userId, year, month)
	if err != nil {
		g.Log().Errorf(ctx, "计算连续签到天数失败: %v", err)
		return err
	}

	// 2. 计算连续签到奖励积分
	// 3. 更新用户积分汇总表和用户积分明细表
	// 如何避免重复发放连续签到奖励? --> 使用 user_monthly_bonus_log 表记录用户指定月份已领取的奖励
	// 2.1 查询用户本月已领取的奖励（避免重复发放）
	var bonusLogs []*entity.UserMonthlyBonusLog
	if err := dao.UserMonthlyBonusLog.Ctx(ctx).
		Where(dao.UserMonthlyBonusLog.Columns().UserId, userId).
		Where(dao.UserMonthlyBonusLog.Columns().YearMonth, fmt.Sprintf("%d%02d", year, month)). // 202505
		Scan(&bonusLogs); err != nil && !errors.Is(err, sql.ErrNoRows) {
		g.Log().Errorf(ctx, "查询用户已领取的奖励失败: %v", err)
		return err
	}

	// 用 BonusType 做去重 key（比用 Description 更稳）
	// 把领取的奖励塞到map中，方便后续判断是否已经发放过奖励
	bonusLogsMap := make(map[ConsecutiveBonusType]bool)
	for _, v := range bonusLogs {
		bonusLogsMap[ConsecutiveBonusType(v.BonusType)] = true
	}

	// 遍历连续签到奖励配置，如果符合条件就发奖励
	for _, rule := range consecutiveBonusRules {
		if maxConsecutive >= rule.TriggerDays && !bonusLogsMap[rule.BonusType] {
			// 发放连续签到奖励积分
			// 更新 user_points 表和 user_points_transactions 表
			if err := s.AddPoints(ctx, &model.PointsTransactionInput{
				UserId: userId,
				Points: rule.Points,
				Desc:   consecutiveBonusNames[rule.BonusType],
				Type:   int(PointsTransactionTypeConsecutive),
			}); err != nil {
				g.Log().Errorf(ctx, "发放连续签到奖励失败: %v", err)
				continue
			}

			// 记录到 user_monthly_bonus_log 表（用于幂等）
			now := time.Now()
			newLog := &entity.UserMonthlyBonusLog{
				UserId:      userId,
				YearMonth:   fmt.Sprintf("%d%02d", year, month),
				Description: consecutiveBonusNames[rule.BonusType],
				BonusType:   int(rule.BonusType),
				CreatedAt:   gtime.NewFromTime(now),
				UpdatedAt:   gtime.NewFromTime(now),
			}

			if _, err := dao.UserMonthlyBonusLog.Ctx(ctx).Insert(newLog); err != nil {
				// 积分已加，连续签到奖励已发，但月度奖励记录插入失败，需要手动处理
				g.Log().Errorf(ctx, "[NEED_HANDLE]插入用户月度奖励记录失败: %v", err)
				continue
			}

			// 更新本地 map，避免本次循环中重复发
			bonusLogsMap[rule.BonusType] = true
		}
	}

	return nil
}

// CalcMonthConsecutiveDays 计算本月最大连续签到天数（含补签）
func (s *Service) CalcMonthConsecutiveDays(ctx context.Context, userId uint64, year, month int) (int, error) {
	// 从用户年度签到记录中取出当月签到 bitmap
	key := fmt.Sprintf(yearSignKeyFormat, userId, year)
	firstOfMonthOffset := getFirstOfMonthOffset(year, month)
	monthDays := getMonthDays(year, month)
	bitWidthType := fmt.Sprintf("u%d", monthDays) // u30/u31

	values, err := s.rc.BitField(ctx, key, "GET", bitWidthType, firstOfMonthOffset).Result()
	if err != nil {
		g.Log().Errorf(ctx, "获取用户签到记录到失败: %v", err)
		return 0, err
	}

	if len(values) == 0 {
		values = []int64{0} // 如果没有查询到，则默认为0
	}

	checkinBitmap := uint64(values[0])
	g.Log().Debugf(ctx, "checkinBitmap: %0b", checkinBitmap)

	// 获取当月补签 bitmap
	retroKey := fmt.Sprintf(monthRetroKeyFormat, userId, year, month)

	// 去 Redis 里 retroKey 这个补签位图，从第 0 位开始，读出 monthDays（比如 30/31）个 bit，打包成一个整数返回
	retroValues, err := s.rc.BitField(ctx, retroKey, "GET", bitWidthType, "#0").Result()
	if err != nil {
		g.Log().Errorf(ctx, "获取用户补签记录失败: %v", err)
		return 0, err
	}

	if len(retroValues) == 0 {
		retroValues = []int64{0} // 没有查询到，则默认为0
	}

	retroBitmap := uint64(retroValues[0])

	// 逻辑或
	bitmap := checkinBitmap | retroBitmap // 合并本月签到和补签数据

	return calcMaxConsecutiveDays(bitmap, monthDays), nil
}

// calcMaxConsecutiveDays 计算最大连续签到天数
func calcMaxConsecutiveDays(bitmap uint64, monthDays int) int {
	// 逐位判断，计算出连续签到天数
	maxCount := 0
	currCount := 0
	for i := range monthDays {
		// 从右向左逐位判断
		checked := (bitmap>>i)&1 == 1
		if checked {
			currCount++
		} else {
			if currCount > maxCount {
				maxCount = currCount
			}
			currCount = 0
		}
	}
	// 循环结束再最后比较一次
	if currCount > maxCount {
		maxCount = currCount
	}
	return maxCount
}

// getFirstOfMonthOffset 获取当月第一天在一年中的偏移量
// 假设 2025-12-01：
// time.Date(2025, 12, 1, ...) 得到 12 月 1 日
// firstOfMonth.YearDay() 得到它是一年中的第几天（2025 不是闰年，12/1 是第 335 天）
// 那么 offset = 335 - 1 = 334
// 注意：这里使用 time.Local，保证和 Daily() 的 time.Now() 同一时区口径
func getFirstOfMonthOffset(year, month int) int {
	// 1. 获取当月第一天
	firstOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	// 2. 计算偏移量
	return firstOfMonth.YearDay() - 1 // offset 从 0 开始
}

// getMonhDays 获取当月天数
// 注意：使用 time.Local，避免时区差异导致月界限异常
func getMonthDays(year, month int) int {
	// 1. 获取当月第一天
	firstOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	// 2. 获取当月最后一天
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	return lastOfMonth.Day() // 返回当月天数
}

// AddPoints 封装“积分变更”的事务逻辑：写明细 + 更新/插入汇总
// input.Points 可正可负：正数=加分，负数=扣分（如需禁止扣分可在下面校验）
func (s *Service) AddPoints(ctx context.Context, input *model.PointsTransactionInput) error {
	if input == nil {
		return errors.New("input is nil")
	}
	if input.UserId == 0 {
		return errors.New("invalid userId")
	}
	if input.Points == 0 {
		return nil // 没有变更直接返回
	}

	now := time.Now()

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1) 查询积分汇总（新用户可能没有记录）
		var userPoint entity.UserPoints
		if err := tx.Model(&entity.UserPoints{}).
			Where(dao.UserPoints.Columns().UserId, input.UserId).
			Scan(&userPoint); err != nil && !errors.Is(err, sql.ErrNoRows) {
			g.Log().Errorf(ctx, "查询用户积分汇总表失败: %v", err)
			return err
		}

		isNew := userPoint.Id == 0
		if isNew {
			userPoint = entity.UserPoints{
				UserId: input.UserId,
			}
		}

		// 2) 计算新余额（可选：不允许余额为负）
		newBalance := userPoint.Points + input.Points
		if newBalance < 0 {
			return errors.New("积分不足")
		}

		// 3) 更新汇总
		userPoint.Points = newBalance
		// PointsTotal 通常表示“累计获得”，一般只在加分时累加
		if input.Points > 0 {
			userPoint.PointsTotal += input.Points
		}

		// 4) 插入积分明细
		newRecord := entity.UserPointsTransactions{
			UserId:          input.UserId,
			PointsChange:    input.Points,
			CurrentBalance:  newBalance,
			TransactionType: input.Type,
			Description:     input.Desc,
			CreatedAt:       gtime.NewFromTime(now),
			UpdatedAt:       gtime.NewFromTime(now),
		}
		if _, err := tx.Model(&entity.UserPointsTransactions{}).Insert(&newRecord); err != nil {
			g.Log().Errorf(ctx, "插入用户积分明细表失败: %v", err)
			return err
		}

		// 5) 更新/插入积分汇总
		if isNew {
			userPoint.CreatedAt = gtime.NewFromTime(now)
			userPoint.UpdatedAt = gtime.NewFromTime(now)
			if _, err := tx.Model(&entity.UserPoints{}).Insert(&userPoint); err != nil {
				g.Log().Errorf(ctx, "插入用户积分汇总表失败: %v", err)
				return err
			}
		} else {
			if _, err := tx.Model(&entity.UserPoints{}).
				Where(dao.UserPoints.Columns().UserId, input.UserId).
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
}
