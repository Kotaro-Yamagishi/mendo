package repository

import (
	"context"

	"mendo/internal/apperrors"
	"mendo/internal/domain"
	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/datasource"
)

// OrderBoardRepository は datasource を使った OrderBoard Projection の永続化実装。
// Order と Kitchen 両方のイベントを横断して1つのビューを作る。
type OrderBoardRepository struct {
	ds datasource.OrderBoardDataSource
}

func NewOrderBoardRepository(ds datasource.OrderBoardDataSource) *OrderBoardRepository {
	return &OrderBoardRepository{ds: ds}
}

// ApplyOrderEvent は Order 集約のイベントを受け取って OrderBoard を更新する。
func (r *OrderBoardRepository) ApplyOrderEvent(ctx context.Context, event domain.Event) error {
	switch e := event.(type) {
	case order.OrderCreated:
		row := &datasource.OrderBoardRow{
			OrderID:       e.GetAggregateID(),
			SeatNo:        e.SeatNo,
			OrderStatus:   "pending",
			CookingStatus: "",
			OrderedAt:     &e.OccurredAt,
		}
		if err := r.ds.UpsertOrderBoardRow(ctx, row); err != nil {
			return apperrors.Infrastructure("注文ボードの保存に失敗", err)
		}

	case order.OrderConfirmed:
		existing, err := r.ds.FindOrderBoardRowByID(ctx, e.GetAggregateID())
		if err != nil {
			return apperrors.Infrastructure("注文ボードの取得に失敗", err)
		}
		if existing == nil {
			return nil
		}
		existing.OrderStatus = "confirmed"
		existing.CookingStatus = "waiting"
		if err := r.ds.UpsertOrderBoardRow(ctx, existing); err != nil {
			return apperrors.Infrastructure("注文ボードの保存に失敗", err)
		}

	case order.OrderCancelled:
		existing, err := r.ds.FindOrderBoardRowByID(ctx, e.GetAggregateID())
		if err != nil {
			return apperrors.Infrastructure("注文ボードの取得に失敗", err)
		}
		if existing == nil {
			return nil
		}
		existing.OrderStatus = "canceled"
		existing.CookingStatus = ""
		if err := r.ds.UpsertOrderBoardRow(ctx, existing); err != nil {
			return apperrors.Infrastructure("注文ボードの保存に失敗", err)
		}
	}
	return nil
}

// ApplyKitchenEvent は Kitchen 集約のイベントを受け取って OrderBoard を更新する。
func (r *OrderBoardRepository) ApplyKitchenEvent(ctx context.Context, event domain.Event) error {
	if e, ok := event.(kitchen.CookingCompleted); ok {
		existing, err := r.ds.FindOrderBoardRowByID(ctx, string(e.OrderID))
		if err != nil {
			return apperrors.Infrastructure("注文ボードの取得に失敗", err)
		}
		if existing == nil {
			return nil
		}
		existing.CookingStatus = "completed"
		cookingAt := e.OccurredAt
		existing.CookingAt = &cookingAt
		if err := r.ds.UpsertOrderBoardRow(ctx, existing); err != nil {
			return apperrors.Infrastructure("注文ボードの保存に失敗", err)
		}
	}
	return nil
}

// FindAll は全 OrderBoardRow を取得して OrderBoardEntry スライスで返す。
func (r *OrderBoardRepository) FindAll(ctx context.Context) ([]OrderBoardEntry, error) {
	rows, err := r.ds.FindAllOrderBoardRows(ctx)
	if err != nil {
		return nil, apperrors.Infrastructure("注文ボード一覧の取得に失敗", err)
	}
	entries := make([]OrderBoardEntry, 0, len(rows))
	for _, row := range rows {
		entry := orderBoardRowToEntry(&row)
		entries = append(entries, entry)
	}
	return entries, nil
}

func orderBoardRowToEntry(row *datasource.OrderBoardRow) OrderBoardEntry {
	entry := OrderBoardEntry{
		OrderID:       row.OrderID,
		SeatNo:        row.SeatNo,
		OrderStatus:   row.OrderStatus,
		CookingStatus: row.CookingStatus,
	}
	if row.OrderedAt != nil {
		entry.OrderedAt = *row.OrderedAt
	}
	if row.CookingAt != nil {
		entry.CookingDoneAt = row.CookingAt
	}
	return entry
}
