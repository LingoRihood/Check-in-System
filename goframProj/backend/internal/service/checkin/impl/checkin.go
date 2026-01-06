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
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/redis/go-redis/v9"
)

// 签到相关业务逻辑的具体实现

const (
	yearSignKeyFormat            = "user:checkins:daily:%d:%d"     // user:checkins:daily:12131321421312:2025
	monthRetroKeyFormat          = "user:checkins:retro:%d:%d%02d" // user:checkins:retro:12131321421231202501
	defaultDailyPoints     int64 = 1                               // 每日签到积分（注意：用 int64，和积分字段统一）
	defaultRetroCostPoints       = 100                             // 补签消耗积分
	maxRetroTimesPerMonth        = 3                               // 单月最多补签次数
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
	PointsTransactionTypeRetro:       "补签%s消耗积分",
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

var (
	ErrInvalidRetroDate = errors.New("补签日期无效")
	ErrChecked          = errors.New("日期已签到")
	ErrRetroNotimes     = errors.New("本月补签次数已用完")
	ErrNoEnoughPoints   = gerror.New("积分不足")
)

type Service struct {
	rc *redis.Client
}

func NewService() *Service {
	return &Service{
		rc: injection.MustInvoke[*redis.Client](), // 从注入器中获取 Redis 客户端实例
	}
}

// Daily 每日签到（先 DB 成功，再 setbit）
// ✅ 先 DB 加分成功
// ✅ 再 Redis 标记已签到
// 这样就不会出现“签了但没加分”
func (s *Service) Daily(ctx context.Context, userId uint64) error {
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

	// 2) 先 DB：发放每日签到积分（事务：明细+汇总）
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

	// ✅ 新增：记录活跃用户（用于定时提醒任务取候选用户）
	// 放在 DB + Redis 都成功之后
	_ = TouchActiveUser(ctx, userId, now)

	// 4) 发送连续签到的奖励积分
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

	// 2.1 查询用户本月已领取的奖励（避免重复发放）
	var bonusLogs []*entity.UserMonthlyBonusLog
	if err := dao.UserMonthlyBonusLog.Ctx(ctx).
		Where(dao.UserMonthlyBonusLog.Columns().UserId, userId).
		Where(dao.UserMonthlyBonusLog.Columns().YearMonth, fmt.Sprintf("%d%02d", year, month)). // 202505
		Scan(&bonusLogs); err != nil && !errors.Is(err, sql.ErrNoRows) {
		g.Log().Errorf(ctx, "查询用户已领取的奖励失败: %v", err)
		return err
	}

	bonusLogsMap := make(map[ConsecutiveBonusType]bool)
	for _, v := range bonusLogs {
		bonusLogsMap[ConsecutiveBonusType(v.BonusType)] = true
	}

	for _, rule := range consecutiveBonusRules {
		if maxConsecutive >= rule.TriggerDays && !bonusLogsMap[rule.BonusType] {
			if err := s.AddPoints(ctx, &model.PointsTransactionInput{
				UserId: userId,
				Points: rule.Points,
				Desc:   consecutiveBonusNames[rule.BonusType],
				Type:   int(PointsTransactionTypeConsecutive),
			}); err != nil {
				g.Log().Errorf(ctx, "发放连续签到奖励失败: %v", err)
				continue
			}

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
				g.Log().Errorf(ctx, "[NEED_HANDLE]插入用户月度奖励记录失败: %v", err)
				continue
			}

			bonusLogsMap[rule.BonusType] = true
		}
	}

	return nil
}

// CalcMonthConsecutiveDays 计算本月最大连续签到天数（含补签）
func (s *Service) CalcMonthConsecutiveDays(ctx context.Context, userId uint64, year, month int) (int, error) {
	monthDays := getMonthDays(year, month)
	checkinBitmap, retroBitmap, err := s.getMonthBitmap(ctx, userId, year, month)
	if err != nil {
		g.Log().Errorf(ctx, "获取用户签到记录失败: %v", err)
		return 0, err
	}
	bitmap := checkinBitmap | retroBitmap
	return calcMaxConsecutiveDays(bitmap, monthDays), nil
}

func calcMaxConsecutiveDays(bitmap uint64, monthDays int) int {
	maxCount := 0
	currCount := 0
	for i := 0; i < monthDays; i++ {
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
	if currCount > maxCount {
		maxCount = currCount
	}
	return maxCount
}

func getFirstOfMonthOffset(year, month int) int {
	firstOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	return firstOfMonth.YearDay() - 1
}

func getMonthDays(year, month int) int {
	firstOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	return lastOfMonth.Day()
}

// AddPoints 封装“积分变更”的事务逻辑：写明细 + 更新/插入汇总
func (s *Service) AddPoints(ctx context.Context, input *model.PointsTransactionInput) error {
	if input == nil {
		return errors.New("input is nil")
	}
	if input.UserId == 0 {
		return errors.New("invalid userId")
	}
	if input.Points == 0 {
		return nil
	}

	now := time.Now()

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
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

		newBalance := userPoint.Points + input.Points
		if newBalance < 0 {
			return errors.New("积分不足")
		}

		userPoint.Points = newBalance
		if input.Points > 0 {
			userPoint.PointsTotal += input.Points
		}

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

// MonthDetail 签到详情
func (s *Service) MonthDetail(ctx context.Context, input *model.MonthDetailInput) (*model.MonthDetailOutput, error) {
	checkinBitmap, retroBitmap, err := s.getMonthBitmap(ctx, input.UserId, input.Year, input.Month)
	if err != nil {
		g.Log().Errorf(ctx, "获取年月bitmap失败: %v", err)
		return nil, err
	}

	g.Log().Debugf(ctx, "--> checkinBitmap: %031b retroBitmap:%031b", checkinBitmap, retroBitmap)
	monthDays := getMonthDays(input.Year, input.Month)
	checkinDays := parseBitmap2Days(checkinBitmap, monthDays)
	retroDays := parseBitmap2Days(retroBitmap, monthDays)

	bitmap := checkinBitmap | retroBitmap
	maxConsecutive := calcMaxConsecutiveDays(bitmap, monthDays)

	remainRetroTimes := maxRetroTimesPerMonth - len(retroDays)

	isCheckedToday, err := s.IsCheckedToday(ctx, input.UserId)
	if err != nil {
		g.Log().Errorf(ctx, "查询当天是否签到失败: %v", err)
		return nil, err
	}

	return &model.MonthDetailOutput{
		CheckedInDays:      checkinDays,
		RetroCheckedInDays: retroDays,
		ConsecutiveDays:    maxConsecutive,
		RemainRetroTimes:   remainRetroTimes,
		IsCheckedInToday:   isCheckedToday,
	}, nil
}

func (s *Service) IsCheckedToday(ctx context.Context, userId uint64) (bool, error) {
	now := time.Now()
	year := now.Year()
	key := fmt.Sprintf(yearSignKeyFormat, userId, year)

	dayOffset := now.YearDay() - 1
	value, err := s.rc.GetBit(ctx, key, int64(dayOffset)).Result()
	if err != nil {
		g.Log().Errorf(ctx, "GetBit 获取当天签到状态失败: %v", err)
		return false, err
	}
	return value == 1, nil
}

func parseBitmap2Days(bitmap uint64, monthDays int) []int {
	days := make([]int, 0)
	for i := 0; i < monthDays; i++ {
		if (bitmap & (1 << (monthDays - 1 - i))) != 0 {
			days = append(days, i+1)
		}
	}
	return days
}

func (s *Service) getMonthBitmap(ctx context.Context, userId uint64, year, month int) (uint64, uint64, error) {
	key := fmt.Sprintf(yearSignKeyFormat, userId, year)
	firstOfMonthOffset := getFirstOfMonthOffset(year, month)
	monthDays := getMonthDays(year, month)
	bitWidthType := fmt.Sprintf("u%d", monthDays)

	values, err := s.rc.BitField(ctx, key, "GET", bitWidthType, firstOfMonthOffset).Result()
	if err != nil {
		g.Log().Errorf(ctx, "获取用户签到记录到失败: %v", err)
		return 0, 0, err
	}
	if len(values) == 0 {
		values = []int64{0}
	}

	checkinBitmap := uint64(values[0])
	g.Log().Debugf(ctx, "checkinBitmap: %0b", checkinBitmap)

	retroKey := fmt.Sprintf(monthRetroKeyFormat, userId, year, month)
	retroValues, err := s.rc.BitField(ctx, retroKey, "GET", bitWidthType, "#0").Result()
	if err != nil {
		g.Log().Errorf(ctx, "获取用户补签记录失败: %v", err)
		return 0, 0, err
	}
	if len(retroValues) == 0 {
		retroValues = []int64{0}
	}

	retroBitmap := uint64(retroValues[0])
	return checkinBitmap, retroBitmap, nil
}

// Retro 根据输入的日期进行补签
func (s *Service) Retro(ctx context.Context, userId uint64, date time.Time) error {
	// 1. 判断补签日期是否有效
	if err := s.checkRetroDate(ctx, userId, date); err != nil {
		return err
	}

	// 2. 执行补签逻辑
	retroKey := fmt.Sprintf(monthRetroKeyFormat, userId, date.Year(), date.Month())
	retroOffset := date.Day() - 1

	err := s.rc.SetBit(ctx, retroKey, int64(retroOffset), 1).Err()
	if err != nil {
		g.Log().Errorf(ctx, "SetBit 设置补签状态失败: %v", err)
		return gerror.NewCode(gcode.CodeInternalError)
	}

	// 2.2 补签消耗积分、增加积分、增加积分记录
	if err := s.retroWithTransaction(ctx, userId, date); err != nil {
		// 如果数据库更新失败，则回滚 Redis 中的补签标识
		rerr := s.rc.SetBit(ctx, retroKey, int64(retroOffset), 0).Err()
		if rerr != nil {
			g.Log().Errorf(ctx, "SetBit 回滚补签状态失败: %v", rerr)
			return gerror.NewCode(gcode.CodeInternalError)
		}
		// ✅ 修复：事务失败必须把错误返回出去
		return err
	}

	// ✅ 新增：补签成功也算活跃用户（用于提醒任务候选集）
	_ = TouchActiveUser(ctx, userId, time.Now())

	// 3. 计算连续签到日期发放连续签到奖励
	return s.updateConsecutiveBonus(ctx, userId, date.Year(), int(date.Month()))
}

// retroWithTransaction 补签逻辑，使用事务保证原子性
func (s *Service) retroWithTransaction(ctx context.Context, userId uint64, date time.Time) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var userPoint entity.UserPoints
		if err := tx.Model(dao.UserPoints.Table()).
			Where(dao.UserPoints.Columns().UserId, userId).
			Scan(&userPoint); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				g.Log().Errorf(ctx, "查询用户积分失败: %v", err)
				return err
			}
			userPoint = entity.UserPoints{
				UserId: userId,
			}
		}

		if userPoint.Points < defaultRetroCostPoints {
			return ErrNoEnoughPoints
		}

		pointsChange := -defaultRetroCostPoints + defaultDailyPoints
		nowPoints := userPoint.Points + int64(pointsChange)
		nowTotalPoints := userPoint.PointsTotal + defaultDailyPoints

		retroCostRecord := entity.UserPointsTransactions{
			UserId:          userId,
			PointsChange:    -defaultRetroCostPoints,
			TransactionType: int(PointsTransactionTypeRetro),
			Description:     fmt.Sprintf(PointsTransactionTypeMsgMap[PointsTransactionTypeRetro], date.Format(time.DateOnly)),
			CurrentBalance:  userPoint.Points - defaultRetroCostPoints,
			CreatedAt:       gtime.NewFromTime(time.Now()),
			UpdatedAt:       gtime.NewFromTime(time.Now()),
		}

		if _, err := tx.Model(dao.UserPointsTransactions.Table()).Insert(&retroCostRecord); err != nil {
			g.Log().Errorf(ctx, "插入补签消耗的积分记录失败: %v", err)
			return err
		}

		checkinBonusRecord := entity.UserPointsTransactions{
			UserId:          userId,
			PointsChange:    defaultDailyPoints,
			TransactionType: int(PointsTransactionTypeDaily),
			Description:     PointsTransactionTypeMsgMap[PointsTransactionTypeDaily],
			CurrentBalance:  nowPoints,
			CreatedAt:       gtime.NewFromTime(time.Now()),
			UpdatedAt:       gtime.NewFromTime(time.Now()),
		}

		if _, err := tx.Model(dao.UserPointsTransactions.Table()).Insert(&checkinBonusRecord); err != nil {
			g.Log().Errorf(ctx, "插入补签奖励积分记录失败: %v", err)
			return err
		}

		userPoint.Points = nowPoints
		userPoint.PointsTotal = nowTotalPoints
		if _, err := tx.Model(dao.UserPoints.Table()).
			Where(dao.UserPoints.Columns().UserId, userId).
			Update(&userPoint); err != nil {
			g.Log().Errorf(ctx, "更新用户积分失败: %v", err)
			return err
		}

		return nil
	})
}

func (s *Service) checkRetroDate(ctx context.Context, userId uint64, date time.Time) error {
	now := time.Now()
	if date.Year() > now.Year() ||
		date.Month() != now.Month() ||
		(date.Year() == now.Year() && date.YearDay() >= now.YearDay()) {
		return ErrInvalidRetroDate
	}

	checkinKey := fmt.Sprintf(yearSignKeyFormat, userId, date.Year())
	yearOffset := date.YearDay() - 1
	checked, err := s.rc.GetBit(ctx, checkinKey, int64(yearOffset)).Result()
	if err != nil {
		g.Log().Errorf(ctx, "GetBit 获取当天签到状态失败: %v", err)
		return err
	}
	if checked == 1 {
		return ErrInvalidRetroDate
	}

	retroKey := fmt.Sprintf(monthRetroKeyFormat, userId, date.Year(), date.Month())
	retroOffset := date.Day() - 1
	retroRet, err := s.rc.GetBit(ctx, retroKey, int64(retroOffset)).Result()
	if err != nil {
		g.Log().Errorf(ctx, "GetBit 获取当天补签状态失败: %v", err)
		return err
	}
	if retroRet == 1 {
		return ErrInvalidRetroDate
	}

	retroCount, err := s.rc.BitCount(ctx, retroKey, nil).Result()
	if err != nil {
		g.Log().Errorf(ctx, "BitCount 获取补签次数失败: %v", err)
		return err
	}
	if retroCount >= maxRetroTimesPerMonth {
		return ErrRetroNotimes
	}
	return nil
}
