package service

import (
	"context"
	"fmt"
	"time"

	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
)

// WaitTimeCalculator は待ち時間を計算するドメインサービス。
// Order と Kitchen の両方のデータを使う。どの集約にも属さない。ステートレス。
type WaitTimeCalculator struct {
	orderReader   order.Reader
	kitchenReader kitchen.Reader
}

func NewWaitTimeCalculator(or order.Reader, kr kitchen.Reader) *WaitTimeCalculator {
	return &WaitTimeCalculator{orderReader: or, kitchenReader: kr}
}

// EstimateWaitTime は現在の待ち時間を推定する。
func (c *WaitTimeCalculator) EstimateWaitTime(ctx context.Context, kitchenID kitchen.KitchenID) (time.Duration, error) {
	// 未完了の注文数を取得（Order 集約のデータ）
	pendingOrders, err := c.orderReader.CountPending(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count pending orders: %w", err)
	}

	// 厨房の調理能力を取得（Kitchen 集約のデータ）
	k, err := c.kitchenReader.FindByID(ctx, kitchenID)
	if err != nil {
		return 0, fmt.Errorf("failed to find kitchen: %w", err)
	}
	capacity := k.CookingCapacity()

	// 計算するだけ。状態を変えない
	if capacity == 0 {
		return 30 * time.Minute, nil // フル稼働時は一律30分
	}
	estimatedMinutes := (pendingOrders / capacity) * 5
	return time.Duration(estimatedMinutes) * time.Minute, nil
}
