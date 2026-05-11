//go:build integration

package mysql_test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:test@tcp(localhost:13306)/mendo_test?charset=utf8mb4&parseTime=true&loc=UTC"
	}

	var err error
	testDB, err = connectWithRetry(dsn, 10, 2*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test DB: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: docker run -d --name mendo-test -p 13306:3306 -e MYSQL_ROOT_PASSWORD=test -e MYSQL_DATABASE=mendo_test mysql:8.0\n")
		os.Exit(1)
	}

	if err := runMigrations(testDB); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

func connectWithRetry(dsn string, maxRetries int, interval time.Duration) (*sql.DB, error) {
	var db *sql.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			time.Sleep(interval)
			continue
		}
		if err = db.Ping(); err == nil {
			return db, nil
		}
		db.Close()
		time.Sleep(interval)
	}
	return nil, fmt.Errorf("failed to connect after %d retries: %w", maxRetries, err)
}

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// filename = /path/to/mendo/internal/infrastructure/datasource/mysql/integration_test.go
	// 4階層上がると mendo/ ルートになる
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "..")
}

func runMigrations(db *sql.DB) error {
	migrationsDir := filepath.Join(projectRoot(), "migrations")
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			// 各行からコメントを除去
			lines := strings.Split(stmt, "\n")
			var cleaned []string
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "--") {
					continue
				}
				cleaned = append(cleaned, line)
			}
			cleanedStmt := strings.TrimSpace(strings.Join(cleaned, "\n"))
			if cleanedStmt == "" {
				continue
			}
			if _, err := db.Exec(cleanedStmt); err != nil {
				return fmt.Errorf("exec migration %s: %w\nStatement: %s", f, err, cleanedStmt)
			}
		}
	}
	return nil
}
