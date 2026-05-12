package repository

import (
	"context"

	"mendo/internal/apperrors"
	"mendo/internal/infrastructure/datasource"
)

// OrderReaderRepository は datasource を使った order.Reader の実装。
// CountPending のみを担当する。軽量なクエリ特化リポジトリ。
type OrderReaderRepository struct {
	ds datasource.OrderProjectionDataSource
}

func NewOrderReaderRepository(ds datasource.OrderProjectionDataSource) *OrderReaderRepository {
	return &OrderReaderRepository{ds: ds}
}

// CountPending は pending 状態の注文数を返す。
// order.Reader を満たす。
func (r *OrderReaderRepository) CountPending(ctx context.Context) (int, error) {
	rows, err := r.ds.FindAllOrderProjections(ctx)
	if err != nil {
		return 0, apperrors.Infrastructure("注文一覧の取得に失敗", err)
	}
	count := 0
	for _, row := range rows {
		if row.Status == "pending" {
			count++
		}
	}
	return count, nil
}
