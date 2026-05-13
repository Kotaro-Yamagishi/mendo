# ログ設計

## 設計方針

- Go 標準の `*slog.Logger` をそのまま DI する。独自 Logger interface は作らない
- domain 層は `slog` を import しない。ログは技術的関心事であり、業務知識ではない
- CorrelationID を context 経由で全レイヤーに伝搬し、全ログに自動付与する
- 環境変数でフォーマット（JSON/Text）とレベル（DEBUG/INFO/WARN/ERROR）を切り替える

### なぜ独自 Logger interface を作らないか

slog は Go 1.21 で標準ライブラリに入った構造化ログパッケージ。`slog.Handler` interface でカスタマイズ可能（出力先、フォーマット、フィールド自動付与）。独自 interface を被せると：

- slog の豊富な API（`With`, `WithGroup`, `LogAttrs`, `LogValuer`）を制限してしまう
- ボイラープレートが増える
- ログライブラリ差し替え予定がなければ YAGNI

domain に Logger interface を置くのは DDD 違反。ログは横断的関心事であり、`Order` や `Price` と同列に置くものではない。

## ログの目的

全てのログは以下の3つのいずれかに属する。「なんとなく出す」は禁止。

| 目的 | 問い | 例 |
|------|------|------|
| 障害調査 | 何が壊れた？なぜ？ | エラー詳細、スタックトレース、入力値 |
| 運用監視 | 今、正常に動いてる？ | リクエスト数、処理時間、キュー滞留 |
| 業務追跡 | あの注文どうなった？ | 注文ID、状態遷移、誰がいつ操作した |

## ログを入れる7つの観点

### 1. システム境界の入出力（最重要）

外部との接点には必ずログを出す。

判定基準: プロセスの外に出るか、外から入ってくるか → YES ならログ必須

### 2. 状態遷移

集約の状態が変わった瞬間。業務追跡の核。ログを出す場所は application 層。

判定基準: ステータスが変わるか → YES ならログ必須

### 3. エラーと異常系

ログは1回だけ。ErrorMiddleware の原則と同じ。

### 4. 非同期処理のライフサイクル

goroutine で動く処理は、起動・停止・処理結果を必ず出す。

### 5. セキュリティイベント

認証失敗、認可チェック、レートリミット等。

### 6. パフォーマンス指標

リクエスト処理時間、スロークエリ検知。

### 7. 冪等性・重複チェック

重複イベント検知、Outbox 配信済みマーク等。

## ログレベル

| レベル | 意味 | 基準 |
|--------|------|------|
| ERROR | 即座に対応が必要 | オンコールが起きるべき事象 |
| WARN | 異常だが即死はしない | 翌営業日に確認すべき事象 |
| INFO | 正常な業務イベント | 運用で「起きたこと」を追跡するため |
| DEBUG | 開発時のみ必要 | 本番では OFF（デフォルト INFO） |

クライアントの入力ミス（400系）は WARN。サーバーが壊れたわけではない。

## 各層のログの書き方

### domain 層: ログを出さない

```go
// ❌ domain で slog を使わない
func (o *Order) Confirm() error {
    slog.Info("confirming order")  // ダメ
    ...
}

// ✅ domain はエラーを返すだけ。ログは上位層の責務
func (o *Order) Confirm() error {
    if o.status != StatusPending {
        return apperrors.Domain(ErrCodeAlreadyConfirmed, "この注文は既に確定されています")
    }
    ...
}
```

### application 層: 状態遷移の成功ログ

「この業務操作が成功した」を記録する。全ての副作用が完了した後、return の直前に配置。

```go
func (uc *ConfirmOrderESUsecase) Execute(ctx context.Context, orderID string) error {
    // 1. ロード → コマンド実行 → 保存 → Publish
    ...

    // 全て成功した後にログ
    slog.InfoContext(ctx, "order confirmed",
        slog.String("order_id", orderID),
    )
    return nil
}
```

```go
func (uc *CreateOrderUsecase) Execute(ctx context.Context, input CreateOrderInput) (string, error) {
    ...

    slog.InfoContext(ctx, "order created",
        slog.String("order_id", orderID),
        slog.Int("seat_no", input.SeatNo),
        slog.Int("item_count", len(input.Items)),
    )
    return orderID, nil
}
```

ルール:
- `slog.InfoContext(ctx, ...)` を使う（CorrelationID が自動付与される）
- エラーパスにはログを入れない（ErrorMiddleware が処理する）
- 業務ルールの判定結果を書かない（domain の責務）

### infrastructure 層: 技術的イベント

logger を DI で受け取り、技術的なイベントをログする。

```go
// EventBus: イベント発行
b.logger.InfoContext(ctx, "event published",
    slog.String("event_type", event.GetEventType()),
    slog.String("aggregate_id", event.GetAggregateID()),
)

// EventBus: リトライ（WARN）
b.logger.WarnContext(ctx, "event handler retry",
    slog.Int("attempt", attempt),
    slog.Int("max_retries", b.maxRetries),
    slog.String("handler", handlerName),
    slog.String("error", err.Error()),
)

// EventBus: DLQ 送り（ERROR）
b.logger.ErrorContext(ctx, "event sent to DLQ",
    slog.String("event_type", eventType),
    slog.String("handler", handlerName),
    slog.String("error", lastErr.Error()),
)

// OutboxRelay: ライフサイクル
r.logger.Info("outbox relay started", slog.String("interval", r.interval.String()))
r.logger.Info("outbox relay delivered", slog.Int("count", len(events)))
r.logger.Error("outbox relay error", "error", err)
```

ルール:
- `*slog.Logger` を struct に持ち、コンストラクタで DI
- goroutine のライフサイクルは必ず起動/停止をログ
- ctx がない場合（goroutine 起動時等）は `logger.Info(...)` でよい

### interface/middleware 層: リクエスト/エラー

```go
// CorrelationMiddleware: リクエスト開始/終了（自動）
logger.InfoContext(ctx, "request started",
    slog.String("method", r.Method),
    slog.String("path", r.URL.Path),
)
logger.InfoContext(ctx, "request completed",
    slog.String("method", r.Method),
    slog.String("path", r.URL.Path),
    slog.Int("status", sw.status),
    slog.Int64("duration_ms", duration.Milliseconds()),
)

// ErrorMiddleware: エラー処理（自動）
// サーバーエラー → logger.ErrorContext
// クライアントエラー → logger.WarnContext
```

ルール:
- handler 自身はログを出さない。ErrorMiddleware に任せる
- handler は `return err` するだけ

### server/subscribers 層: イベント購読トレース

```go
// イベント受信時
slog.InfoContext(ctx, "event published",
    slog.String("event_type", order.EventTypeOrderConfirmed),
    slog.String("action", "convert to public event"),
    slog.String("order_id", publicEvent.OrderID),
)

// Saga 補償
slog.InfoContext(ctx, "saga compensation",
    slog.String("event_type", kitchen.EventTypeCookingRejected),
    slog.String("action", "cancel order"),
    slog.String("order_id", rejected.OrderID),
)
```

ルール:
- `slog.InfoContext(ctx, ...)` でグローバル slog を使う（subscribers は関数なので DI できない）
- イベント型は定数を使う

### Projection: DEBUG レベル

```go
slog.Debug("projection insert",
    slog.String("order_id", e.GetAggregateID()),
    slog.Int("seat_no", e.SeatNo),
    slog.String("status", "pending"),
)
```

ルール:
- 本番では出力されない（デフォルト INFO）
- 開発時に `LOG_LEVEL=DEBUG` で有効化

## CorrelationID

- HTTP 境界で生成（`X-Correlation-ID` ヘッダ or 新規 UUID）
- `context.WithValue` で全レイヤーに伝搬
- `correlationHandler`（カスタム slog.Handler）が全ログに自動付与
- DomainEvent の CorrelationID と一致させる

## DI の流れ

```
cmd/main.go
  logging.NewLogger() → *slog.Logger 生成
  slog.SetDefault(logger)
  di.InitializeApp(..., logger)
    ↓
  Wire が *slog.Logger を各コンストラクタに注入
    → WatermillEventBus
    → RelayService
    → Worker
    → App.Logger
  server.Run(app, bus)
    app.Logger を使用
    RegisterRoutes(mux, app, ..., logger)
    CorrelationMiddleware(mux, logger)
```

## 環境変数

| 変数 | 値 | デフォルト | 説明 |
|------|------|-----------|------|
| LOG_FORMAT | json / text | text | JSON は本番用、Text は開発用 |
| LOG_LEVEL | DEBUG / INFO / WARN / ERROR | INFO | 出力する最低レベル |

## Linter

`sloglint` で slog の使い方を統一する。

- `kv-only: true`: `slog.String("key", val)` 形式（typed attr）に統一

## 実装後チェックリスト

```
□ 全ての HTTP エンドポイントにリクエスト/レスポンスログがあるか
□ 全てのイベント発行/受信にログがあるか
□ 全ての goroutine に起動/停止ログがあるか
□ 全てのエラーパスにログがあるか（ただし1回だけ）
□ 状態遷移する全てのユースケースにログがあるか
□ 外部システム呼び出し（DB, API）の失敗にログがあるか
□ リトライ処理にリトライログがあるか
□ fmt.Printf / fmt.Println / log.Printf が 0 件か
□ domain 層で slog を import していないか
```
