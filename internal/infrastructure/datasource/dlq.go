package datasource

import "context"

// DLQDataSource は dlq テーブルへのアクセス IF。
// 引数・戻り値はすべてプリミティブ型または DTO。ドメイン型を使わない。
type DLQDataSource interface {
	// InsertDeadLetterRow はリトライ失敗したイベントを dlq テーブルに保存する。
	InsertDeadLetterRow(ctx context.Context, row *DeadLetterRow) error

	// FindAllDeadLetterRows は dlq テーブルの全行を返す。
	FindAllDeadLetterRows(ctx context.Context) ([]DeadLetterRow, error)

	// FindDeadLetterRowByID は id を指定して DeadLetterRow を返す。
	// 見つからない場合は nil, nil を返す。
	FindDeadLetterRowByID(ctx context.Context, id string) (*DeadLetterRow, error)

	// DeleteDeadLetterRow は指定した id の行を削除する。
	DeleteDeadLetterRow(ctx context.Context, id string) error
}
