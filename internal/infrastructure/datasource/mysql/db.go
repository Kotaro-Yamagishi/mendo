package mysql

import (
	"context"
	"database/sql"
)

// txKey はコンテキストにトランザクションを格納するキー。
// infrastructure/mysql パッケージと同じ仕組みをこのパッケージ内でも保持する。
type txKey struct{}

// conn はクエリを実行できる共通インターフェース。
// *sql.DB と *sql.Tx の両方が満たす。
type conn interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// getConn はコンテキストにトランザクションがあればそれを返し、
// なければ *sql.DB を返す。トランザクション透過的に SQL を実行するために使う。
func getConn(ctx context.Context, db *sql.DB) conn {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return db
}
