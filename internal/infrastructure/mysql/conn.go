package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Config はMySQL接続設定。
// 環境変数やconfigファイルから読み込む想定。
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// DefaultConfig はローカル開発用のデフォルト設定。
// docker-compose.yml の設定と合わせる。
func DefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "root",
		Database: "mendo",
	}
}

// NewConnection はMySQL接続を作成する。
func NewConnection(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=UTC",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}

	return db, nil
}
