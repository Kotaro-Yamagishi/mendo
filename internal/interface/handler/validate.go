package handler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"mendo/internal/apperrors"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	// JSON タグ名をフィールド名として使う（"seat_no" 等）
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// validateInput は struct のバリデーションを実行し、エラーを ValidationWithFields に変換する。
func validateInput(input any) error {
	err := validate.Struct(input)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return apperrors.Validation("VALIDATION_ERROR", "入力内容に誤りがあります")
	}

	var fieldErrors []apperrors.FieldError
	for _, fe := range validationErrors {
		fieldErrors = append(fieldErrors, apperrors.FieldError{
			Field:   formatFieldName(fe.Namespace(), input),
			Message: formatMessage(fe),
		})
	}

	return apperrors.ValidationWithFields(fieldErrors)
}

// formatFieldName は Namespace からトップレベルの struct 名を除去し、
// 配列インデックスを含むフィールド名を生成する。
// 例: "createOrderRequest.items[0].menu_id" → "items[0].menu_id"
func formatFieldName(namespace string, input any) string {
	t := reflect.TypeOf(input)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	prefix := t.Name() + "."
	name := strings.TrimPrefix(namespace, prefix)
	return name
}

// formatMessage はバリデーションタグに応じた日本語メッセージを返す。
func formatMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "必須項目です"
	case "min":
		if fe.Kind() == reflect.Slice {
			return fmt.Sprintf("%s個以上必要です", fe.Param())
		}
		return fmt.Sprintf("%s以上の値を指定してください", fe.Param())
	case "max":
		if fe.Kind() == reflect.Slice {
			return fmt.Sprintf("%s個以下にしてください", fe.Param())
		}
		return fmt.Sprintf("%s以下の値を指定してください", fe.Param())
	case "gte":
		return fmt.Sprintf("%s以上の値を指定してください", fe.Param())
	default:
		return "不正な値です"
	}
}
