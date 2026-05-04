# スキーマ設計

mendo の各テーブルのスキーマ設計。InMemory 実装だが、本番 DB にした場合の構造を示す。

## 全サービス共通のテーブル

以下のテーブルは集約の種類やサービスに関係なく同じ構造になる。

### events テーブル（EventStore）

イベントソーシングの根幹。append-only（INSERT のみ）。

| カラム | 型 | 説明 |
|-------|---|------|
| event_id | TEXT PK | イベントの一意ID |
| aggregate_id | TEXT NOT NULL | どの集約のイベントか |
| aggregate_type | TEXT NOT NULL | 集約の種類（Order 等） |
| event_type | TEXT NOT NULL | OrderConfirmed 等 |
| version | INTEGER NOT NULL | 集約内の連番 |
| payload | JSONB NOT NULL | イベントのデータ |
| created_at | TIMESTAMP NOT NULL | 作成日時 |

- UNIQUE(aggregate_id, version) で同じ集約の同じ版を防止
- UPDATE / DELETE しない

対応コード: `infrastructure/eventstore/inmemory.go`

### snapshots テーブル（未実装）

集約の復元を高速化するキャッシュ。Read Model とは別物。

| カラム | 型 | 説明 |
|-------|---|------|
| aggregate_id | TEXT | どの集約か |
| aggregate_type | TEXT | 集約の種類 |
| version | INTEGER | 何件目のイベントまでの状態か |
| state | JSONB | 集約の状態を丸ごとJSON化 |
| created_at | TIMESTAMP | 作成日時 |

- PK: (aggregate_id, version)
- 集約の復元時: snapshot をロード → snapshot.version 以降の events だけ再生
- N件ごと（例: 100件）に自動作成する運用
- mendo ではイベント数が少ないため未実装

### outbox テーブル（Outbox Pattern）

イベントの確実な配信を保証する送信箱。

| カラム | 型 | 説明 |
|-------|---|------|
| id | TEXT PK | 送信箱エントリのID |
| event_type | TEXT NOT NULL | イベント種別 |
| payload | JSONB NOT NULL | 公開イベントのデータ |
| delivered | BOOLEAN DEFAULT FALSE | 配信済みフラグ |
| created_at | TIMESTAMP NOT NULL | 作成日時 |

- 集約の Save と同じトランザクションで Store する
- Relay Service が定期ポーリングして EventBus / Kafka に配信
- 配信成功後に delivered = true にマーク

対応コード: `infrastructure/outbox/inmemory.go`, `infrastructure/outbox/relay.go`

### dlq テーブル（Dead Letter Queue）

リトライしても失敗したイベントの退避先。

| カラム | 型 | 説明 |
|-------|---|------|
| id | TEXT PK | DLQエントリのID |
| event_type | TEXT NOT NULL | 元のイベント種別 |
| payload | JSONB NOT NULL | 元のイベントデータ |
| error | TEXT NOT NULL | 失敗理由 |
| fail_count | INTEGER NOT NULL | リトライ回数 |
| handler_name | TEXT NOT NULL | 失敗したハンドラ名 |
| last_fail_at | TIMESTAMP NOT NULL | 最後に失敗した日時 |

対応コード: `infrastructure/dlq/inmemory.go`

## サービスごとに異なるテーブル（Projection）

Projection は画面や用途に合わせた構造のため、サービスごとに異なる。

### order_projections（注文サービスの Read Model）

注文一覧 API 用。Order 集約のイベントから構築。

| カラム | 型 | 説明 |
|-------|---|------|
| order_id | TEXT PK | 注文ID |
| seat_no | INTEGER | 座席番号 |
| items | JSONB | 注文品目 |
| total | INTEGER | 合計金額 |
| status | TEXT | pending / confirmed / cancelled |
| created_at | TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | 更新日時 |

対応コード: `infrastructure/projection/order_state_store.go`

### order_board（厨房の注文ボード）

厨房スタッフ用。Order + Kitchen の2つの集約のイベントを統合した横断ビュー。

| カラム | 型 | 説明 |
|-------|---|------|
| order_id | TEXT PK | 注文ID |
| seat_no | INTEGER | 座席番号 |
| order_status | TEXT | Order のイベントから |
| cooking_status | TEXT | Kitchen のイベントから |
| ordered_at | TIMESTAMP | 注文日時 |
| cooking_at | TIMESTAMP | 調理完了日時 |

対応コード: `infrastructure/projection/order_board.go`

## テーブルの関係

```
events テーブル（イベントの記録）
  ↓ リプレイで集約を復元（書き込み用）
snapshots テーブル（復元の高速化キャッシュ）

events テーブル
  ↓ イベント配信（subscriber 経由）
order_projections テーブル（読み取りビュー）
order_board テーブル（横断読み取りビュー）

events テーブル
  ↓ 同じトランザクションで保存
outbox テーブル
  ↓ Relay Service が配信
EventBus / Kafka

EventBus
  ↓ リトライ全失敗
dlq テーブル
```
