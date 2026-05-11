package repository

import (
	"context"
	"fmt"
	"time"

	"mendo/internal/domain/specialorder"
	"mendo/internal/infrastructure/datasource"
)

// SpecialOrderRepository は datasource を使った SpecialOrder の永続化実装。
// specialorder.Reader と specialorder.Writer を実装する。
type SpecialOrderRepository struct {
	ds datasource.SpecialOrderDataSource
}

func NewSpecialOrderRepository(ds datasource.SpecialOrderDataSource) *SpecialOrderRepository {
	return &SpecialOrderRepository{ds: ds}
}

// FindByID は SpecialOrderRow を取得してドメインモデルを復元する。
func (r *SpecialOrderRepository) FindByID(ctx context.Context, id specialorder.SpecialOrderID) (*specialorder.SpecialOrder, error) {
	row, err := r.ds.FindSpecialOrderByID(ctx, id.String())
	if err != nil {
		return nil, fmt.Errorf("SpecialOrderRepository.FindByID: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("special order not found: %s", id)
	}
	return rowToSpecialOrder(row), nil
}

// Save は SpecialOrder を SpecialOrderRow に変換して永続化する。
func (r *SpecialOrderRepository) Save(ctx context.Context, so *specialorder.SpecialOrder) error {
	now := time.Now().UTC()
	row := &datasource.SpecialOrderRow{
		ID:            so.ID().String(),
		OrderID:       so.OrderID(),
		MenuName:      so.MenuName(),
		Status:        int(so.Status()),
		SuggestedMenu: so.SuggestedMenu(),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := r.ds.UpsertSpecialOrder(ctx, row); err != nil {
		return fmt.Errorf("SpecialOrderRepository.Save: %w", err)
	}
	return nil
}

func rowToSpecialOrder(row *datasource.SpecialOrderRow) *specialorder.SpecialOrder {
	return specialorder.ReconstructSpecialOrder(
		specialorder.SpecialOrderID(row.ID),
		row.OrderID,
		row.MenuName,
		specialorder.SpecialOrderStatus(row.Status),
		row.SuggestedMenu,
	)
}
