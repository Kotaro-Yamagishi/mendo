package handler_test

import (
	"log/slog"
	"net/http"

	"mendo/internal/interface/handler"
)

// testLogger はテスト用のサイレントロガー。
var testLogger = slog.Default()

// wrap は AppHandlerFunc を ErrorMiddleware でラップして http.HandlerFunc に変換する。
// テストで直接ハンドラを呼び出す際に使う。
func wrap(h handler.AppHandlerFunc) http.HandlerFunc {
	return handler.ErrorMiddleware(h, testLogger)
}
