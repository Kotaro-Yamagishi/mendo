# Event Sourcing

`domain/order/` にイベントソーシング版の Order 集約を実装。従来版は `_archive/order_traditional/` に保存。

## 従来版 vs ES 版

```
従来版（_archive/order_traditional/）:
  Order テーブルに最新状態を UPDATE
  → 「なぜこの状態か」は追えない

ES 版（domain/order/）:
  events テーブルにイベントを INSERT only
  → 全履歴が残る。イベント列を再生すれば任意の時点の状態を復元可能
```

## ES のデータフロー

```
コマンド実行:
  order.Confirm()
    → 業務ルールチェック
    → OrderConfirmed イベントを生成
    → apply() で自分の状態に反映
    → uncommittedEvents に追加

保存:
  eventStore.Save(order.UncommittedEvents())
    → events テーブルに INSERT（上書きしない）

ロード:
  events := eventStore.Load(aggregateID)
  order := ReconstructFromEvents(events)
    → イベントを順に Apply して現在の状態を復元
```

## データモデル（テーブル設計）

**events テーブル**（INSERT only。全集約共通。真実の源）

| event_id | aggregate_id | event_type | event_data | version |
|----------|-------------|------------|------------|---------|
| evt-1 | order-001 | order.created | {"seat_no":3} | 0 |
| evt-2 | order-001 | order.item_added | {"menu_id":"ramen-01"} | 1 |
| evt-3 | order-001 | order.confirmed | {} | 2 |

**order_projections テーブル**（UPDATE。読み取り用キャッシュ。消えても events から再生成可能）

| order_id | seat_no | status | item_count |
|----------|---------|--------|-----------|
| order-001 | 3 | confirmed | 1 |

## Projection（リードモデル）

同じイベント列から目的別のモデルを生成:

```
events テーブル
  ↓ Load
イベント列
  ├→ OrderStateProjection      管理画面表示用（現在の状態）
  └→ OrderAnalyticsProjection  分析用（注文数、トッピング数、確定までの時間）
```

## 書き込みと読み取りの分離（CQRS）

```
書き込み: usecase → eventStore.Save() → events テーブルに INSERT
                 → Publish → subscriber → Projection テーブルを UPDATE

読み取り: usecase → projectionReader.FindAll() → Projection テーブルから SELECT
                 → events テーブルは触らない
```

## Projection の配置ルール

```
このリードモデルは1つの集約のイベントだけで作れる？
  ├─ Yes → domain/{集約名}/projection.go
  └─ No（複数集約のイベントが必要）
       → infrastructure/projection/
```

具体例:

```
domain/order/projection.go       ← Order のイベントだけ使う
domain/kitchen/projection.go     ← Kitchen のイベントだけ使う（あれば）

infrastructure/projection/order_board.go
  → Order のイベント（OrderCreated, OrderConfirmed）
  + Kitchen のイベント（CookingCompleted）
  = 「注文状況ボード」というリードモデルを生成
  用途: 店舗モニターに「3番テーブル: 濃厚鶏白湯 → 調理完了」と表示
```

## ディレクトリ構成

```
domain/
├── order/            # ES 版（イベント列として保存）
│   ├── order.go      # 集約（Apply + コマンド + 業務ルール）
│   ├── events.go     # イベント定義
│   ├── values.go     # 値オブジェクト（OrderID, SeatNumber）
│   └── projection.go # リードモデル（Projection）
├── eventstore.go     # EventStore IF（domain 層で定義）

_archive/
└── order_traditional/ # 従来版（状態をそのまま保存）

infrastructure/
├── eventstore/
│   └── inmemory.go # EventStore 実装（学習用インメモリ）

application/
├── confirm_order_es.go  # ES 版ユースケース
```

## 関連ドキュメント

- [Business Flow](business-flow.md) — 注文から提供までの全体フロー
- [Event Design](event-design.md) — イベントの構造とメタデータ
- [Schema Design](schema-design.md) — 全テーブルのスキーマ設計
