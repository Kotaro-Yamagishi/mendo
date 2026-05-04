package datasource

import "context"

// MenuDataSource は oc_menus テーブルへのアクセス IF。
// 引数・戻り値はすべてプリミティブ型または DTO。ドメイン型を使わない。
type MenuDataSource interface {
	// FindMenuByID は menu_id を指定して MenuRow を返す。
	// 見つからない場合は nil, nil を返す。
	FindMenuByID(ctx context.Context, id string) (*MenuRow, error)

	// FindAllMenus は全メニューを返す。
	FindAllMenus(ctx context.Context) ([]MenuRow, error)

	// InsertMenu は新しいメニューを INSERT する。
	InsertMenu(ctx context.Context, row *MenuRow) error

	// UpdateMenuAvailability は指定した menu_id の availability を更新する。
	UpdateMenuAvailability(ctx context.Context, id string, available bool) error
}
