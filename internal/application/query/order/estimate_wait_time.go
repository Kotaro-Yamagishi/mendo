package order

import (
	"context"
	"time"

	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/service"
)

// EstimateWaitTimeUsecase は待ち時間推定のユースケース。
// ドメインサービス（WaitTimeCalculator）を呼び出す。
type EstimateWaitTimeUsecase struct {
	calculator *service.WaitTimeCalculator
	kitchenID  kitchen.KitchenID
}

func NewEstimateWaitTimeUsecase(calc *service.WaitTimeCalculator, id kitchen.KitchenID) *EstimateWaitTimeUsecase {
	return &EstimateWaitTimeUsecase{calculator: calc, kitchenID: id}
}

func (uc *EstimateWaitTimeUsecase) Execute(ctx context.Context) (time.Duration, error) {
	d, err := uc.calculator.EstimateWaitTime(ctx, uc.kitchenID)
	if err != nil {
		return 0, err
	}
	return d, nil
}
