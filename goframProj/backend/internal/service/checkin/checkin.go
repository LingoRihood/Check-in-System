package checkin

import (
	"backend/internal/model"
	"context"
)

type Service interface {
	Daily(ctx context.Context, userID uint64) error                                                   // 每日签到
	MonthDetail(ctx context.Context, input *model.MonthDetailInput) (*model.MonthDetailOutput, error) // 签到详情
}
