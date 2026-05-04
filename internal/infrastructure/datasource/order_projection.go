package datasource

import "context"

// OrderProjectionDataSource は oc_order_projections テーブルへのアクセス IF。
// 引数・戻り値はすべてプリミティブ型または DTO。ドメイン型を使わない。
type OrderProjectionDataSource interface {
	// FindOrderProjectionByID は order_id を指定して OrderProjectionRow を返す。
	// 見つからない場合は nil, nil を返す。
	FindOrderProjectionByID(ctx context.Context, orderID string) (*OrderProjectionRow, error)

	// FindAllOrderProjections は全 Order の Projection を返す。
	FindAllOrderProjections(ctx context.Context) ([]OrderProjectionRow, error)

	// UpsertOrderProjection は Order Projection を INSERT OR UPDATE する。
	UpsertOrderProjection(ctx context.Context, row *OrderProjectionRow) error
}
