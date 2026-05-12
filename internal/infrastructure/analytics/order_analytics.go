package analytics

import (
	"context"
	"fmt"
	"time"

	"mendo/internal/apperrors"
	"mendo/internal/domain"
	"mendo/internal/domain/order"
)

// OrderAnalyticsProjection は注文の分析モデル（OLAP）を構築する Projection。
// OrderConfirmed イベントを受けて fact_orders に INSERT する。
// CQRS の Read 側として、業務イベントから分析用の星形スキーマを構築する。
//
// データメッシュの観点:
//   - この Projection は「注文BCの分析データプロダクト」
//   - 注文BC のチームが業務系と分析系の両方を管理する
//   - 公開インターフェースとして分析クエリ用のビューを提供する
type OrderAnalyticsProjection struct {
	facts []OrderFact
}

// OrderFact は事実テーブルの1行に対応する。
type OrderFact struct {
	FactID    string
	OrderID   string
	DateKey   int
	MenuKey   string
	SeatKey   int
	Amount    int
	Quantity  int
	OrderedAt time.Time
}

// CookingFact は調理事実テーブルの1行に対応する。
type CookingFact struct {
	FactID      string
	OrderID     string
	DateKey     int
	MenuKey     string
	StartedAt   time.Time
	CompletedAt time.Time
	DurationSec int
}

func NewOrderAnalyticsProjection() *OrderAnalyticsProjection {
	return &OrderAnalyticsProjection{}
}

// HandleOrderEvent は注文イベントを受けて分析モデルに投影する。
// OrderConfirmed イベントから fact_orders を生成する。
func (p *OrderAnalyticsProjection) HandleOrderEvent(_ context.Context, event domain.Event) error {
	switch event.GetEventType() {
	case order.EventTypeOrderConfirmed:
		return p.applyOrderConfirmed(event)
	default:
		// 分析に不要なイベントはスキップ
		return nil
	}
}

func (p *OrderAnalyticsProjection) applyOrderConfirmed(event domain.Event) error {
	confirmed, ok := event.(*order.OrderConfirmed)
	if !ok {
		return apperrors.Infrastructure("予期しないイベント型", fmt.Errorf("expected *order.OrderConfirmed, got %T", event))
	}

	now := time.Now()
	dateKey := toDateKey(now)

	for i, item := range confirmed.Items {
		fact := OrderFact{
			// Items[i] で一意にする。同一メニューが複数ある場合も index で区別する
			FactID:  fmt.Sprintf("%s-%s-%d", confirmed.GetAggregateID(), item.MenuID, i),
			OrderID: confirmed.GetAggregateID(),
			DateKey: dateKey,
			MenuKey: item.MenuID,
			SeatKey: confirmed.SeatNo,
			// ConfirmedItem に Price/Quantity フィールドがないため、
			// 金額は dim_menu 結合時に計算する。ここは数量 1 で記録する
			Amount:    0,
			Quantity:  1,
			OrderedAt: now,
		}
		p.facts = append(p.facts, fact)
	}

	return nil
}

// GetFacts は蓄積された事実を返す（テスト・デバッグ用）。
func (p *OrderAnalyticsProjection) GetFacts() []OrderFact {
	return p.facts
}

func toDateKey(t time.Time) int {
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
}
