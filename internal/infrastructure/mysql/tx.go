package mysql

import (
	"context"
	"database/sql"
	"fmt"
)

// txKey はコンテキストにトランザクションを格納するキー。
type txKey struct{}

// conn はクエリを実行できる共通インターフェース。
// *sql.DB と *sql.Tx の両方が満たす。
type conn interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// GetConn はコンテキストにトランザクションがあればそれを返し、
// なければ *sql.DB を返す。トランザクション透過的に SQL を実行するために使う。
func GetConn(ctx context.Context, db *sql.DB) conn {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return db
}

// MySQLTransactionManager は MySQL のトランザクション管理。
// domain.TransactionManager を満たす。
type MySQLTransactionManager struct {
	db *sql.DB
}

func NewMySQLTransactionManager(db *sql.DB) *MySQLTransactionManager {
	return &MySQLTransactionManager{db: db}
}

func (tm *MySQLTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("transaction rolled back: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
