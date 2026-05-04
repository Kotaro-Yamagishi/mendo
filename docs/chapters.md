# チャプターマッピング — 各章の概念が mendo のどこに実装されているか

## 第1章: 事業活動を分析する（業務領域の3分類）

| 概念 | mendo での実装 |
|------|--------------|
| 事業領域 | ラーメンチェーン「麺道」 |
| 中核の業務領域 | Order BC（注文管理） |
| 補完の業務領域 | Kitchen BC（調理管理）、Menu BC（メニュー管理） |

## 第2章: 業務知識を発見する（ユビキタス言語）

| 概念 | mendo での実装 |
|------|--------------|
| ユビキタス言語 | domain/ 配下の型名・メソッド名が業務用語そのまま |
| 用語集 | `.claude/rules/ddd.md` で命名ルールを定義 |

## 第3章: 区切られた文脈（Bounded Context）

| 概念 | mendo での実装 |
|------|--------------|
| Bounded Context | Order BC, Kitchen BC, Menu BC が独立 |
| BC 間の疎結合 | EventBus 経由のイベント連携。直接 import しない |

## 第4章: 連携パターン

| 概念 | mendo での実装 |
|------|--------------|
| 共用サービス（OHS） | EventBus が公開イベントを配信 |
| モデル変換層（ACL） | `ToPublicConfirmed()` 等で内部→公開に変換 |

## 第5章: 単純な業務ロジック

| 概念 | mendo での実装 |
|------|--------------|
| トランザクションスクリプト | mendo では不採用（ドメインモデルを使用） |
| アクティブレコード | mendo では不採用 |

## 第6章: ドメインモデル

| 概念 | mendo での実装 |
|------|--------------|
| 集約のルート | `domain/order/order.go` の Order struct |
| 内部エンティティ | `domain/kitchen/task.go` の CookingTask |
| 値オブジェクト | `domain/order/values.go`（OrderID, SeatNumber） |
| ドメインイベント（内部） | `domain/order/events.go`（OrderConfirmed 等） |
| ドメインサービス | `domain/service/wait_time_calculator.go` |
| 業務ルール | `Order.Confirm()` 内の制約チェック |
| リポジトリ IF | `domain/order/repository.go`, `domain/kitchen/repository.go` |

## 第7章: イベントソーシング

| 概念 | mendo での実装 |
|------|--------------|
| Event Sourcing | Order BC で採用。`domain/order/order.go` の Apply パターン |
| EventStore IF | `domain/eventstore.go` |
| EventStore 実装 | `infrastructure/eventstore/inmemory.go` |
| Projection（リードモデル） | `domain/order/projection.go`（構造体）+ `infrastructure/projection/order_state_store.go`（実装） |
| Apply | `order.ReconstructFromEvents()` でイベント列から集約を復元 |
| State Sourcing | Kitchen BC, Menu BC で採用（従来型の Repository） |

## 第8章: 技術方式

| 概念 | mendo での実装 |
|------|--------------|
| ポートとアダプター（依存関係逆転） | domain に IF、infrastructure に実装 |
| CQRS | `application/command/`（書き込み）と `application/query/`（読み取り）に分離 |
| Projection テーブル | `infrastructure/projection/order_state_store.go` |
| 横断リードモデル | `infrastructure/projection/order_board.go` |

## 第9章: 通信

| 概念 | mendo での実装 |
|------|--------------|
| 内部イベント | `domain/order/events.go`（OrderConfirmed） |
| 公開イベント（BC 間の契約） | `internal/domain/contract/order.go`（OrderConfirmedPublic 等） |
| 内部→公開変換関数 | `domain/order/public_events.go`（ToPublicConfirmed 等） |
| モデル変換 | `server/subscribers_order.go` で `ToPublicConfirmed()` を呼び contract/ の型に変換 |
| 送信箱（Outbox） | `domain/outbox.go`（IF）+ `infrastructure/outbox/inmemory.go`（実装） |
| リレーサービス | `infrastructure/outbox/relay.go` |
| サーガ | `server/subscribers_order.go` + `server/subscribers_kitchen.go` のイベント連鎖 |
| 補償アクション | Kitchen が CookingRejected を発行 → subscriber が Order.Cancel() を実行 |
| DLQ（Dead Letter Queue） | `domain/dlq.go`（IF）+ `infrastructure/dlq/inmemory.go`（実装）|
| DLQ リトライ | `infrastructure/eventbus/watermill.go` で maxRetries 回リトライ後 DLQ へ |
| DLQ 管理 API | `GET /admin/dlq`（一覧）+ `POST /admin/dlq/{id}/retry`（再実行） |
| プロセスマネージャー | `domain/specialorder/` — 状態遷移を持つ集約として実装 |
| プロセスマネージャー API | `POST /special-orders`（作成）+ `/approve`（承認）+ `/reject`（却下）+ `/resubmit`（再申請） |

## 第10章: 設計の経験則

| 概念 | mendo での実装 |
|------|--------------|
| 判断マトリックス | 業務領域の分類 → 実装パターン → アーキテクチャ → テスト方針 |
| トランザクションスクリプト | `internal/closing/` — CloseShop() 関数1つで完結。モデルなし |
| アクティブレコード | `internal/staff/` — Staff 構造体に Validate + Store が同居。usecase 層なし |
| ドメインモデル | `internal/domain/order/`, `kitchen/`, `menu/` — 集約 + Repository IF 分離 |
| イベント駆動式ドメインモデル | `internal/domain/order/` — EventStore + Apply + CQRS |
| パターンの混在 | 同じプロジェクト内で4パターンが共存 |
| バッチユースケース（非同期） | `domain/import/` + `infrastructure/importworker/` — 非同期キュー + チャンク並列処理 + プログレス管理 |

## golangci-lint（横断）

| 概念 | mendo での実装 |
|------|--------------|
| DDD 依存方向チェック | `.golangci.yml` の depguard ルール |
| ハードコーディング禁止 | `.golangci.yml` の forbidigo ルール |
| コード品質 | 27 linter 有効化 |

## Wire DI（横断）

| 概念 | mendo での実装 |
|------|--------------|
| 依存関係の組立 | `di/wire.go`, `di/providers.go`, `di/wire_gen.go` |
| App 構造体 | `di/app.go` |
