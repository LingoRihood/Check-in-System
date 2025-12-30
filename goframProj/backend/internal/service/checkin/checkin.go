package checkin

import (
	"context"
)

type Service interface {
	Daily(ctx context.Context, userID uint64) error // 每日签到
}
