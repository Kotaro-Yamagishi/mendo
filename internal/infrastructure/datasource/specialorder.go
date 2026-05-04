package datasource

import "context"

// SpecialOrderDataSource は sc_special_orders テーブルへのアクセス IF。
// 引数・戻り値はすべてプリミティブ型または DTO。ドメイン型を使わない。
type SpecialOrderDataSource interface {
	// FindSpecialOrderByID は id を指定して SpecialOrderRow を返す。
	// 見つからない場合は nil, nil を返す。
	FindSpecialOrderByID(ctx context.Context, id string) (*SpecialOrderRow, error)

	// UpsertSpecialOrder は SpecialOrder を INSERT OR UPDATE する。
	UpsertSpecialOrder(ctx context.Context, row *SpecialOrderRow) error
}
