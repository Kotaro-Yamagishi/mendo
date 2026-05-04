package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"mendo/internal/domain"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/datasource"
)

// OrderProjectionRepository は datasource を使った Order Projection の永続化実装。
// order.ProjectionReader と order.ProjectionWriter を実装する。
type OrderProjectionRepository struct {
	ds datasource.OrderProjectionDataSource
}

func NewOrderProjectionRepository(ds datasource.OrderProjectionDataSource) *OrderProjectionRepository {
	return &OrderProjectionRepository{ds: ds}
}

// HandleEvent はドメインイベントを受け取って Projection を更新する。
// order.ProjectionWriter を満たす。
func (r *OrderProjectionRepository) HandleEvent(ctx context.Context, event domain.Event) error {
	switch e := event.(type) {
	case order.OrderCreated:
		items, _ := json.Marshal([]order.Item{})
		row := &datasource.OrderProjectionRow{
			OrderID:   e.GetAggregateID(),
			SeatNo:    e.SeatNo,
			Items:     string(items),
			Total:     0,
			Status:    "pending",
			CreatedAt: e.OccurredAt,
			UpdatedAt: e.OccurredAt,
		}
		if err := r.ds.UpsertOrderProjection(ctx, row); err != nil {
			return fmt.Errorf("OrderProjectionRepository.HandleEvent OrderCreated: %w", err)
		}

	case order.ItemAdded:
		existing, err := r.ds.FindOrderProjectionByID(ctx, e.GetAggregateID())
		if err != nil {
			return fmt.Errorf("OrderProjectionRepository.HandleEvent ItemAdded find: %w", err)
		}
		if existing == nil {
			return nil
		}
		var items []order.Item
		_ = json.Unmarshal([]byte(existing.Items), &items)
		items = append(items, order.Item{
			MenuID:   e.MenuID,
			Toppings: e.Toppings,
			Hardness: e.Hardness,
		})
		itemsJSON, _ := json.Marshal(items)
		existing.Items = string(itemsJSON)
		existing.UpdatedAt = e.OccurredAt
		if err := r.ds.UpsertOrderProjection(ctx, existing); err != nil {
			return fmt.Errorf("OrderProjectionRepository.HandleEvent ItemAdded upsert: %w", err)
		}

	case order.OrderConfirmed:
		existing, err := r.ds.FindOrderProjectionByID(ctx, e.GetAggregateID())
		if err != nil {
			return fmt.Errorf("OrderProjectionRepository.HandleEvent OrderConfirmed find: %w", err)
		}
		if existing == nil {
			return nil
		}
		existing.Status = "confirmed"
		existing.UpdatedAt = e.OccurredAt
		if err := r.ds.UpsertOrderProjection(ctx, existing); err != nil {
			return fmt.Errorf("OrderProjectionRepository.HandleEvent OrderConfirmed upsert: %w", err)
		}

	case order.OrderCancelled:
		existing, err := r.ds.FindOrderProjectionByID(ctx, e.GetAggregateID())
		if err != nil {
			return fmt.Errorf("OrderProjectionRepository.HandleEvent OrderCancelled find: %w", err)
		}
		if existing == nil {
			return nil
		}
		existing.Status = "canceled"
		existing.UpdatedAt = e.OccurredAt
		if err := r.ds.UpsertOrderProjection(ctx, existing); err != nil {
			return fmt.Errorf("OrderProjectionRepository.HandleEvent OrderCancelled upsert: %w", err)
		}
	}
	return nil
}

// FindByID は OrderProjectionRow を取得して OrderStateRow を返す。
// order.ProjectionReader を満たす。
func (r *OrderProjectionRepository) FindByID(ctx context.Context, orderID string) (*order.OrderStateRow, error) {
	row, err := r.ds.FindOrderProjectionByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("OrderProjectionRepository.FindByID: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("order projection not found: %s", orderID)
	}
	return projectionRowToStateRow(row), nil
}

// FindAll は全 OrderProjectionRow を取得して OrderStateRow スライスを返す。
// order.ProjectionReader を満たす。
func (r *OrderProjectionRepository) FindAll(ctx context.Context) ([]order.OrderStateRow, error) {
	rows, err := r.ds.FindAllOrderProjections(ctx)
	if err != nil {
		return nil, fmt.Errorf("OrderProjectionRepository.FindAll: %w", err)
	}
	result := make([]order.OrderStateRow, 0, len(rows))
	for i := range rows {
		result = append(result, *projectionRowToStateRow(&rows[i]))
	}
	return result, nil
}

func projectionRowToStateRow(row *datasource.OrderProjectionRow) *order.OrderStateRow {
	var items []order.Item
	_ = json.Unmarshal([]byte(row.Items), &items)
	return &order.OrderStateRow{
		OrderID:   row.OrderID,
		SeatNo:    row.SeatNo,
		Status:    row.Status,
		ItemCount: len(items),
	}
}

