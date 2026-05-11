.PHONY: test test-unit test-integration test-all cover lint build

# ユニットテストのみ（デフォルト。高速）
test:
	go test -race -short ./...

# ユニットテストのみ（test と同じ。明示的な名前）
test-unit:
	go test -race ./...

# 統合テスト（Docker が必要）
# 事前に: docker run -d --name mendo-test -p 13306:3306 -e MYSQL_ROOT_PASSWORD=test -e MYSQL_DATABASE=mendo_test mysql:8.0
test-integration:
	go test -race -tags=integration ./internal/infrastructure/datasource/mysql/...

# 全テスト（ユニット + 統合）
test-all: test-unit test-integration

# カバレッジレポート
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# カバレッジ率のみ表示
cover-func:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

# Lint
lint:
	golangci-lint run ./...

# ビルド
build:
	go build ./...
