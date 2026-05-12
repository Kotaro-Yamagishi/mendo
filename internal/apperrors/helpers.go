package apperrors

import "errors"

// IsCode はエラーが指定のコードを持つ AppError かを判定する。
// テストで使う。
func IsCode(err error, code string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// GetAppError はエラーから AppError を取り出す。
// 見つからなければ nil を返す。
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}
