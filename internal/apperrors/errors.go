package apperrors

import "fmt"

// Category はエラーの分類。HTTP ステータスに1:1対応。
type Category int

const (
	CategoryValidation     Category = iota // 400
	CategoryUnauthorized                   // 401
	CategoryForbidden                      // 403
	CategoryNotFound                       // 404
	CategoryTimeout                        // 408
	CategoryConflict                       // 409
	CategoryTooLarge                       // 413
	CategoryDomain                         // 422
	CategoryInfrastructure                 // 500
	CategoryBadGateway                     // 502
	CategoryUnavailable                    // 503
)

// HTTPStatus は Category を HTTP ステータスコードに変換する。
func (c Category) HTTPStatus() int {
	switch c {
	case CategoryValidation:
		return 400
	case CategoryUnauthorized:
		return 401
	case CategoryForbidden:
		return 403
	case CategoryNotFound:
		return 404
	case CategoryTimeout:
		return 408
	case CategoryConflict:
		return 409
	case CategoryTooLarge:
		return 413
	case CategoryDomain:
		return 422
	case CategoryBadGateway:
		return 502
	case CategoryUnavailable:
		return 503
	default:
		return 500
	}
}

// IsClientError は Category がクライアント側のエラーかを判定。
// クライアントエラーの場合、メッセージをユーザーに返す。
// サーバーエラーの場合、メッセージを隠蔽する。
func (c Category) IsClientError() bool {
	return c < CategoryInfrastructure
}

// AppError はアプリケーション全体で使うエラー型。
type AppError struct {
	Category Category       // エラーの分類
	Code     string         // 機械可読コード（API利用者向け）
	Message  string         // 人間可読メッセージ（ユーザー向け）
	Cause    error          // 元エラー（ログ用。ユーザーには見せない）
	Details  map[string]any // 構造化コンテキスト（ログ用）
	CorrelationID string         // 分散トレーシング用
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithDetails はエラーにコンテキストを追加する。
func (e *AppError) WithDetails(details map[string]any) *AppError {
	e.Details = details
	return e
}

// WithCause は元エラーを追加する。
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}
