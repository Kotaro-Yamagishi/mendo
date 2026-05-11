package mysql

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

// compactJSON は MySQL の JSON カラムから返る空白付き JSON を空白なしに正規化する。
// MySQL は JSON カラムの値を {"key": "value"} 形式で返すため、
// テスト等で {"key":"value"} と比較する際に不一致が起きる。
func compactJSON(buf *bytes.Buffer, src []byte) error {
	if err := json.Compact(buf, src); err != nil {
		return fmt.Errorf("json compact: %w", err)
	}
	return nil
}
