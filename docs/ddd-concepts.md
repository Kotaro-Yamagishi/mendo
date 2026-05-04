# DDD Concepts

DDD の概念とこのプロジェクトのコードの対応表。

## 概念とコードの対応

| 概念 | どこにいる | 何をする | このプロジェクトでの例 |
|------|----------|---------|---------------------|
| **業務ルール** | 集約の中 | 「やっていい？ダメ？」の判定 | 品切れチェック、トッピング上限、フル稼働チェック |
| **ドメインイベント** | 集約が発行 | 「何が起きた」を外に伝える | OrderConfirmed, CookingCompleted |
| **ドメインサービス** | domain/service/ | 集約をまたぐ計算。ステートレス | WaitTimeCalculator |
| **ユースケース** | application/ | ロード→実行→保存→イベント配信 | ConfirmOrderUsecase, StartCookingUsecase |
| **集約のルート** | domain/{集約名}/ | 業務ルール + 状態変更 + イベント発行 | Order, Menu, Kitchen |
| **内部エンティティ** | 集約の中 | ルート経由でのみ操作 | OrderItem, CookingTask |
| **値オブジェクト** | values.go | IDなし、不変、バリデーション内蔵 | Topping, CookingNote, Price |
| **リポジトリ IF** | domain/{集約名}/ | 集約の保存・取得の抽象 | order.Reader, kitchen.Writer |

## 集約ルートのルール

- 外部からの**唯一のアクセス入口**
- 内部エンティティを直接操作させない。ルート経由でのみ変更
- フィールドは private（小文字）。アクセサ経由で公開
- 業務ルールはコマンドメソッド内に書く

## ドメインイベントのルール

- 過去形で命名する（`OrderConfirmed`, `CookingCompleted`）
- 集約内では `append` で溜めるだけ。Publish はユースケースの責務
- イベント名（EventType）と集約種別（AggregateType）は定数で定義

```go
// 正しい
const AggregateTypeOrder = "Order"
const EventTypeOrderConfirmed = "order.confirmed"

// ダメ（文字列リテラル直書き）
domain.NewDomainEvent(id, "Order", "order.confirmed", corr)
```

## 値オブジェクトのルール

- 不変（Immutable）。作成後に変更しない
- バリデーションはコンストラクタ（`New*`, `Parse*`）に内蔵
- string や int をそのまま使わず意味のある型にする

```go
// ダメ
func AddItem(menuID string, topping string) error

// 正しい
func AddItem(menuID MenuID, topping Topping) error
```

## ドメインサービスのルール

- `domain/service/` に配置。特定の集約に属さない
- ステートレス（自分は状態を持たない）
- 複数の集約やリポジトリのデータを使う計算に限定

```go
// WaitTimeCalculator: 集約をまたぐ計算の典型例
func (c *WaitTimeCalculator) EstimateWaitTime(kitchenID KitchenID) (time.Duration, error) {
    pendingCount := c.orderReader.CountPending()   // Order 集約のデータ
    kitchen := c.kitchenReader.FindByID(kitchenID) // Kitchen 集約のデータ
    // 状態を変えない。計算するだけ。
}
```

## BC 間の公開イベント（contract/）

- BC 間のデータ交換は `internal/domain/contract/` の型のみ使う
- 他の BC の `domain/` を直接 import しない
- 変換関数（`ToPublicXxx`）は発行元の BC の `domain/` に置く

```
Order BC が Kitchen BC にイベントを送る場合:
  1. Order BC: OrderConfirmed（内部イベント）を生成
  2. Order BC: contract.OrderConfirmedPublic に変換（変換関数は Order BC 側に置く）
  3. EventBus: contract.OrderConfirmedPublic を配信
  4. Kitchen BC: contract.OrderConfirmedPublic を受信
  Kitchen BC は Order の内部イベントを知らない
```

## 4パターンの使い分け

| パターン | どこで使う | 選ぶ理由 |
|---------|----------|---------|
| Transaction Script | closing/ | ロジックがシンプル。集約・イベント不要 |
| Active Record | staff/ | CRUD 中心。業務ルールが薄い。補完領域 |
| Domain Model | kitchen/, menu/ | 業務ルールが複雑。集約間連携がある |
| Event Sourcing | order/ | 履歴追跡が必要。監査ログが欲しい |

## 関連ドキュメント

- [Business Flow](business-flow.md) — パターンが実際にどう動くか
- [Event Design](event-design.md) — ドメインイベントの構造とメタデータ
- [Architecture](architecture.md) — レイヤー構成と各層の責務
