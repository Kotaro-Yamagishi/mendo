package apperrors

import "fmt"

// --- Domain (422) ---

func Domain(code, message string) *AppError {
	return &AppError{Category: CategoryDomain, Code: code, Message: message}
}

// --- NotFound (404) ---

func NotFound(resource, id string) *AppError {
	return &AppError{
		Category: CategoryNotFound,
		Code:     "NOT_FOUND",
		Message:  fmt.Sprintf("%s not found: %s", resource, id),
		Details:  map[string]any{"resource": resource, "id": id},
	}
}

// --- Conflict (409) ---

func Conflict(code, message string) *AppError {
	return &AppError{Category: CategoryConflict, Code: code, Message: message}
}

// --- Validation (400) ---

func Validation(code, message string) *AppError {
	return &AppError{Category: CategoryValidation, Code: code, Message: message}
}

// FieldError はバリデーションエラーの個別フィールド。
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func ValidationWithFields(fields []FieldError) *AppError {
	return &AppError{
		Category: CategoryValidation,
		Code:     "VALIDATION_ERROR",
		Message:  "入力内容に誤りがあります",
		Details:  map[string]any{"fields": fields},
	}
}

// --- Unauthorized (401) ---

func Unauthorized(message string) *AppError {
	return &AppError{Category: CategoryUnauthorized, Code: "UNAUTHORIZED", Message: message}
}

// --- Forbidden (403) ---

func Forbidden(message string) *AppError {
	return &AppError{Category: CategoryForbidden, Code: "FORBIDDEN", Message: message}
}

// --- Infrastructure (500) ---

func Infrastructure(message string, cause error) *AppError {
	return &AppError{
		Category: CategoryInfrastructure,
		Code:     "INTERNAL_ERROR",
		Message:  message,
		Cause:    cause,
	}
}

// --- BadGateway (502) ---

func BadGateway(message string, cause error) *AppError {
	return &AppError{
		Category: CategoryBadGateway,
		Code:     "BAD_GATEWAY",
		Message:  message,
		Cause:    cause,
	}
}

// --- Unavailable (503) ---

func Unavailable(message string) *AppError {
	return &AppError{
		Category: CategoryUnavailable,
		Code:     "SERVICE_UNAVAILABLE",
		Message:  message,
	}
}
