package impl

import (
	"backend/utility/injection"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/redis/go-redis/v9"
)

//go:embed remind.lua
var remindScript string

// 活跃签到用户集合：member=userId(字符串), score=最后一次签到时间戳(Unix秒)
const activeUsersZSetKey = "checkin:active_users"

// ZSet 分页拉取大小
const zrangePageSize int64 = 2000

var (
	shaOnce sync.Once
	shaVal  string
	shaErr  error
)

// TouchActiveUser ✅ 需要在“签到/补签成功后”调用，用于让提醒任务知道哪些用户是候选用户
func TouchActiveUser(ctx context.Context, userID uint64, t time.Time) error {
	rc := injection.MustInvoke[*redis.Client]()
	// ZSET 结构：
	// member：userId（字符串）
	// score：最后一次签到时间（Unix 秒）
	return rc.ZAdd(ctx, activeUsersZSetKey, redis.Z{
		Score:  float64(t.Unix()),
		Member: strconv.FormatUint(userID, 10),
	}).Err()
}

func loadLuaSha(ctx context.Context, rc *redis.Client) (string, error) {
	shaOnce.Do(func() {
		sha, err := rc.ScriptLoad(ctx, remindScript).Result()
		if err != nil {
			shaErr = err
			return
		}
		shaVal = sha
	})
	return shaVal, shaErr
}

func reloadLuaSha(ctx context.Context, rc *redis.Client) (string, error) {
	sha, err := rc.ScriptLoad(ctx, remindScript).Result()
	if err != nil {
		return "", err
	}
	shaVal = sha
	shaErr = nil
	return sha, nil
}

func isNoScriptErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToUpper(err.Error()), "NOSCRIPT")
}

// fetchCandidateUserIDs 从活跃 ZSET 拉候选用户：最近 remindThreshold 天内有签到行为，且最后一次签到时间 < 今天开始（意味着“今天没签过”）
func fetchCandidateUserIDs(ctx context.Context, rc *redis.Client, now time.Time, remindThreshold int) ([]uint64, error) {
	// 今天 00:00:00（本地时区）
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 只扫最近 remindThreshold 天内的活跃用户；再加 1 天缓冲避免边界问题
	start := todayStart.Add(-time.Duration(remindThreshold+1) * 24 * time.Hour).Unix()
	end := todayStart.Add(-1 * time.Second).Unix() // < todayStart

	var out []uint64
	var offset int64 = 0

	for {
		members, err := rc.ZRangeByScore(ctx, activeUsersZSetKey, &redis.ZRangeBy{
			Min:    strconv.FormatInt(start, 10),
			Max:    strconv.FormatInt(end, 10),
			Offset: offset,
			Count:  zrangePageSize,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil, nil
			}
			return nil, err
		}
		if len(members) == 0 {
			break
		}

		for _, m := range members {
			id, err := strconv.ParseUint(m, 10, 64)
			if err != nil {
				// 跳过脏数据
				continue
			}
			out = append(out, id)
		}

		if int64(len(members)) < zrangePageSize {
			break
		}
		offset += zrangePageSize
	}

	return out, nil
}

// CheckAndNotify 检查签到并发送通知：连续签到了今天之前的 remindThreshold 天，但今天没签
func CheckAndNotify(ctx context.Context, remindThreshold int) error {
	if remindThreshold <= 0 {
		return fmt.Errorf("invalid remindThreshold: %d", remindThreshold)
	}

	now := time.Now()
	dayOfYearOffset := now.YearDay() - 1 // 0-based

	// 拿 Redis 客户端
	rc := injection.MustInvoke[*redis.Client]()

	// 1) 获取候选用户（避免全量遍历）
	userIDs, err := fetchCandidateUserIDs(ctx, rc, now, remindThreshold)
	if err != nil {
		return fmt.Errorf("fetchCandidateUserIDs err: %w", err)
	}
	if len(userIDs) == 0 {
		return nil
	}

	// 2) 预加载脚本 SHA（缓存一次）
	sha, err := loadLuaSha(ctx, rc)
	if err != nil {
		return fmt.Errorf("ScriptLoad err: %w", err)
	}

	// 3) 遍历候选用户，用 Lua 精准判断
	for _, userID := range userIDs {
		key := fmt.Sprintf(yearSignKeyFormat, userID, now.Year())

		result, evalErr := rc.EvalSha(ctx, sha, []string{key}, dayOfYearOffset, remindThreshold).Int()
		if evalErr != nil && isNoScriptErr(evalErr) {
			// Redis 重启/flush 后脚本丢失：重载并重试一次
			newSha, rerr := reloadLuaSha(ctx, rc)
			if rerr != nil {
				return fmt.Errorf("reload script err: %w", rerr)
			}
			sha = newSha
			result, evalErr = rc.EvalSha(ctx, sha, []string{key}, dayOfYearOffset, remindThreshold).Int()
		}
		if evalErr != nil {
			// 单个用户失败不影响整体（你也可以选择直接 return）
			fmt.Printf("EvalSha err user=%d: %v\n", userID, evalErr)
			continue
		}

		if result == 1 {
			fmt.Printf("用户%d需要发送断签提醒\n", userID)
			// TODO: 发送到消息队列，执行后续的推送逻辑（APP Push 或 短信等）
		}
	}

	return nil
}
