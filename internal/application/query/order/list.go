package order

import (
	"context"

	"mendo/internal/domain/order"
)

// ListOrdersUsecase は注文一覧取得のユースケース。
// Projection（リードモデル）から読み取る。events テーブルは触らない。
type ListOrdersUsecase struct {
	projectionReader order.ProjectionReader
}

func NewListOrdersUsecase(pr order.ProjectionReader) *ListOrdersUsecase {
	return &ListOrdersUsecase{projectionReader: pr}
}

func (uc *ListOrdersUsecase) Execute(ctx context.Context) ([]order.OrderStateRow, error) {
	rows, err := uc.projectionReader.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (uc *ListOrdersUsecase) FindByID(ctx context.Context, orderID string) (*order.OrderStateRow, error) {
	row, err := uc.projectionReader.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return row, nil
}
