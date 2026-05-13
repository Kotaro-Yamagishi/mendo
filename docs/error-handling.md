# Error Handling

## 概要

mendo のエラーハンドリングは `AppError` 型を中心に設計されている。エラーは Category で分類され、HTTP ステータスコードに自動マッピングされる。

## AppError

```go
type AppError struct {
    Category Category       // エラーの分類（HTTP ステータスに対応）
    Code     string         // 機械可読コード（API利用者がswitch等で判定）
    Message  string         // 人間可読メッセージ（ユーザーに表示）
    Cause    error          // 元エラー（ログ用。ユーザーには見せない）
    Details  map[string]any // 構造化コンテキスト（ログ用）
    CorrelationID string         // 分散トレーシング用
}
```

## Category と HTTP ステータス

| Category | HTTP Status | 用途 | メッセージ |
|----------|------------|------|----------|
| CategoryValidation | 400 | 入力値の不正 | ユーザーに返す |
| CategoryUnauthorized | 401 | 認証失敗 | ユーザーに返す |
| CategoryForbidden | 403 | 権限不足 | ユーザーに返す |
| CategoryNotFound | 404 | リソース未発見 | ユーザーに返す |
| CategoryTimeout | 408 | タイムアウト | ユーザーに返す |
| CategoryConflict | 409 | 状態の競合 | ユーザーに返す |
| CategoryTooLarge | 413 | サイズ超過 | ユーザーに返す |
| CategoryDomain | 422 | 業務ルール違反 | ユーザーに返す |
| CategoryInfrastructure | 500 | 技術的障害 | **隠蔽** |
| CategoryBadGateway | 502 | 外部サービス障害 | **隠蔽** |
| CategoryUnavailable | 503 | サービス停止 | **隠蔽** |

Category >= Infrastructure のエラーはメッセージを隠蔽し、ログにのみ詳細を出力する。

## レイヤーごとの責務

### domain 層
- `apperrors.Domain(code, message)` を返す
- エラーコード定数は各集約の `errors.go` に定義
- 業務の言葉でメッセージを書く

```go
// domain/order/errors.go
const ErrCodeAlreadyConfirmed = "ORDER_ALREADY_CONFIRMED"

// domain/order/order.go
func (o *Order) Confirm() error {
    if o.status != StatusPending {
        return apperrors.Domain(ErrCodeAlreadyConfirmed, "この注文は既に確定されています")
    }
}
```

### application 層
- domain エラーはそのまま返す（ラップしない）
- リソース未発見: `apperrors.NotFound("order", id)`
- インフラエラー: `apperrors.Infrastructure("保存に失敗", err)`

```go
func (uc *ConfirmOrderUsecase) Execute(ctx context.Context, id string) error {
    events, err := uc.eventStore.Load(ctx, id)
    if err != nil {
        return apperrors.Infrastructure("注文の読み込みに失敗", err)
    }
    if len(events) == 0 {
        return apperrors.NotFound("order", id)
    }
    o := order.ReconstructFromEvents(events)
    if err := o.Confirm(); err != nil {
        return err // domain エラーはそのまま
    }
    // ...
}
```

### infrastructure 層
- `apperrors.Infrastructure()` を返す
- SQL エラーの詳細は Cause に入れる

### handler 層
- `AppHandlerFunc`（error を返す handler）を使う
- `ErrorMiddleware` が自動でエラーを変換する
- handler はステータスコードを意識しない

```go
func (h *OrderHandler) HandleConfirm(w http.ResponseWriter, r *http.Request) error {
    if err := h.confirmUC.Execute(ctx, orderID); err != nil {
        return err // ErrorMiddleware が処理
    }
    writeSuccess(w, 200, data)
    return nil
}
```

## ErrorMiddleware

全エンドポイントに適用される middleware:

```
handler が error を返す
  ↓
ErrorMiddleware がキャッチ
  ↓
errors.As で AppError を取り出す
  ↓
Category.IsClientError() で判定
  ├→ クライアントエラー: メッセージをそのまま返す（WARN ログ）
  └→ サーバーエラー: "internal server error" に隠蔽（ERROR ログ）
  ↓
CorrelationID を付与してレスポンス
```

## API レスポンス形式

### 成功
```json
{"data": {...}}
```

### クライアントエラー（400-422）
```json
{
    "error": {
        "code": "ORDER_ALREADY_CONFIRMED",
        "message": "この注文は既に確定されています",
        "trace_id": "abc-123-def"
    }
}
```

### バリデーションエラー（400 + fields）
```json
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "入力内容に誤りがあります",
        "trace_id": "abc-123-def",
        "fields": [
            {"field": "name", "message": "名前は必須です"},
            {"field": "email", "message": "メール形式が不正です"}
        ]
    }
}
```

### サーバーエラー（500）
```json
{
    "error": {
        "code": "INTERNAL_ERROR",
        "message": "internal server error",
        "trace_id": "abc-123-def"
    }
}
```

## エラーコード一覧

### 共通
| Code | Category | 説明 |
|------|----------|------|
| NOT_FOUND | 404 | リソース未発見 |
| VALIDATION_ERROR | 400 | 入力値不正 |
| INVALID_REQUEST_BODY | 400 | リクエスト形式不正 |
| UNAUTHORIZED | 401 | 認証失敗 |
| FORBIDDEN | 403 | 権限不足 |
| INTERNAL_ERROR | 500 | 技術的障害 |

### Order 集約
| Code | Category | 説明 |
|------|----------|------|
| ORDER_ALREADY_CONFIRMED | 422 | 注文は既に確定済み |
| ORDER_EMPTY_ITEMS | 422 | 注文品目が空 |
| ORDER_TOO_MANY_TOPPINGS | 422 | トッピング3つ超過 |
| ORDER_MENU_SOLD_OUT | 422 | メニューが品切れ |
| ORDER_NOT_PENDING | 422 | 注文が pending 状態でない |
| ORDER_NOT_CONFIRMED | 422 | 注文が confirmed 状態でない |

### Kitchen 集約
| Code | Category | 説明 |
|------|----------|------|
| KITCHEN_FULL | 422 | 厨房がフル稼働 |
| KITCHEN_TASK_NOT_FOUND | 422 | 調理タスク未発見 |
| KITCHEN_ALREADY_COOKED | 422 | 調理済み |

### SpecialOrder 集約
| Code | Category | 説明 |
|------|----------|------|
| SPECIALORDER_INVALID_STATUS | 422 | 不正な状態遷移 |
| SPECIALORDER_ALREADY_APPROVED | 422 | 既に承認済み |
| SPECIALORDER_NOT_PENDING | 422 | pending でない |

### Menu 集約
| Code | Category | 説明 |
|------|----------|------|
| MENU_INVALID_NAME | 422 | メニュー名が不正 |
| MENU_INVALID_PRICE | 422 | 価格が不正 |

## テストでのエラー判定

```go
// エラーコードで判定（推奨）
assert.True(t, apperrors.IsCode(err, order.ErrCodeAlreadyConfirmed))

// AppError を取り出して詳細を検証
appErr := apperrors.GetAppError(err)
assert.Equal(t, apperrors.CategoryDomain, appErr.Category)
assert.Equal(t, "ORDER_ALREADY_CONFIRMED", appErr.Code)
```

## エラーは1回だけ handle する

```
NG: application 層でログを出して、handler でもログを出す
OK: application 層は return するだけ。handler（middleware）で1回だけログ + レスポンス
```
