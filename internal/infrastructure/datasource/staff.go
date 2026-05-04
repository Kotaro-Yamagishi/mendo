package datasource

import "context"

// StaffDataSource は staffs テーブルへのアクセス IF。
// アクティブレコードパターンの Staff に対応する datasource 層の IF。
// 引数・戻り値はすべてプリミティブ型または DTO。ドメイン型を使わない。
type StaffDataSource interface {
	// FindStaffByID は id を指定して StaffRow を返す。
	// 見つからない場合は nil, nil を返す。
	FindStaffByID(ctx context.Context, id string) (*StaffRow, error)

	// FindAllStaffs は全スタッフを返す。
	FindAllStaffs(ctx context.Context) ([]StaffRow, error)

	// UpsertStaff はスタッフを INSERT OR UPDATE する。
	// ID が空文字の場合は INSERT として扱い、呼び出し元が採番済み ID をセットして渡す。
	UpsertStaff(ctx context.Context, row *StaffRow) error

	// DeleteStaff は指定した id のスタッフを削除する。
	DeleteStaff(ctx context.Context, id string) error
}
