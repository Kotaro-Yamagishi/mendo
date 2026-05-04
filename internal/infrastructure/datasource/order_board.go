package datasource

import "context"

// OrderBoardDataSource は kc_order_board テーブルへのアクセス IF。
// Order と Kitchen の両方のイベントから構築される横断 Projection。
// 引数・戻り値はすべてプリミティブ型または DTO。ドメイン型を使わない。
type OrderBoardDataSource interface {
	// FindOrderBoardRowByID は order_id を指定して OrderBoardRow を返す。
	// 見つからない場合は nil, nil を返す。
	FindOrderBoardRowByID(ctx context.Context, orderID string) (*OrderBoardRow, error)

	// FindAllOrderBoardRows は kc_order_board テーブルの全行を返す。
	FindAllOrderBoardRows(ctx context.Context) ([]OrderBoardRow, error)

	// UpsertOrderBoardRow は OrderBoard エントリを INSERT OR UPDATE する。
	UpsertOrderBoardRow(ctx context.Context, row *OrderBoardRow) error
}
