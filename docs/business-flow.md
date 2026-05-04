# Business Flow

注文から提供までの全体フロー。イベント駆動で集約が連携する様子を図解。

## 概要

```
お客さん → 注文作成 → 注文確定 → 調理開始 → 調理完了 → 提供
              ↑                    ↑
          メニュー参照          イベント自動連携
                                （Order → Kitchen）
```

## イベント連携図

```
┌─ 注文BC ──────────────────────────┐
│                                    │
│  お客さん                           │
│    ↓                               │
│  [注文を確定する]                    │
│    ↓                               │
│  Order 集約                         │
│    ↓                               │
│  OrderConfirmed イベント             │
│    ↓ contract/ で公開イベントに変換   │
│  OrderConfirmedPublic               │
│                                    │
└────────────┬───────────────────────┘
             ↓ EventBus（非同期）
┌─ 厨房BC ──┴───────────────────────┐
│                                    │
│  [自動: 調理タスク作成] ← ポリシー    │
│    ↓                               │
│  Kitchen 集約                       │
│    ├→ 成功: 調理開始                 │
│    │    ↓                          │
│    │  厨房スタッフ                   │
│    │    ↓                          │
│    │  [調理を完了する]               │
│    │    ↓                          │
│    │  CookingCompleted              │
│    │                               │
│    └→ 失敗: キャパオーバー           │
│         ↓                          │
│       CookingRejected               │
│                                    │
└────────────┬───────────────────────┘
             ↓ 補償アクション
┌─ 注文BC ──┴───────────────────────┐
│  [注文をキャンセルする] ← 自動       │
│    ↓                               │
│  OrderCancelled                     │
└────────────────────────────────────┘
```

## 特別注文フロー（プロセスマネージャー）

```
┌─ 特別注文BC ─────────────────────────────┐
│                                           │
│  お客さん                                  │
│    ↓                                      │
│  [特別注文を作成する]                       │
│    ↓                                      │
│  SpecialOrder 集約（状態: Pending）          │
│    ↓                                      │
│  [特別注文一覧] ← ReadModel                 │
│    ↓                                      │
│  店長                                      │
│    ├→ [承認する]                            │
│    │    ↓                                  │
│    │  SpecialOrder（状態: Approved）          │
│    │    ↓                                  │
│    │  CookingDispatched                     │
│    │    ↓ EventBus                         │
│    │  Kitchen 集約へ調理指示                  │
│    │                                       │
│    └→ [拒否する]                            │
│         ↓                                  │
│       SpecialOrder（状態: Rejected）          │
│         ↓                                  │
│       [拒否通知] ← ReadModel                 │
│         ↓                                  │
│       お客さん → [別メニューで再申請]          │
│                                           │
└───────────────────────────────────────────┘
```

## API Endpoints

| Method | Path | Handler | 説明 |
|--------|------|---------|------|
| POST | /orders | OrderHandler.HandleCreate | 注文作成 |
| POST | /orders/{id}/confirm | OrderHandler.HandleConfirm | 注文確定 → 調理開始（自動） |
| POST | /orders/{id}/cancel | OrderHandler.HandleCancel | 注文キャンセル |
| GET | /orders | OrderHandler.HandleList | 注文一覧（Projection） |
| GET | /orders/{id} | OrderHandler.HandleGetByID | 注文詳細（Projection） |
| GET | /wait-time | OrderHandler.HandleWaitTime | 待ち時間推定（ドメインサービス） |
| POST | /kitchen/complete | KitchenHandler.HandleCompleteCooking | 調理完了 |
| POST | /menus/{id}/soldout | MenuHandler.HandleSoldOut | メニュー品切れ |
| GET | /staff | StaffHandler.HandleList | スタッフ一覧（Active Record） |
| GET | /staff/{id} | StaffHandler.HandleGetByID | スタッフ詳細（Active Record） |
| POST | /staff | StaffHandler.HandleCreate | スタッフ作成（Active Record） |
| DELETE | /staff/{id} | StaffHandler.HandleDelete | スタッフ削除 |
| POST | /admin/close-shop | ClosingHandler | 閉店処理（Transaction Script） |
| POST | /admin/import/menus | ImportHandler | メニュー一括取込（非同期） |
| GET | /admin/import/{id}/status | ImportHandler | 取込進捗確認 |
| GET | /admin/dlq | DLQHandler.HandleList | DLQ 一覧 |
| POST | /admin/dlq/{id}/retry | DLQHandler.HandleRetry | DLQ 手動リトライ |

## 全体フロー: 注文から提供まで（コードパス）

```
1. お客さんが食券を買う
   POST /orders → OrderHandler.HandleCreate
     → CreateOrderUsecase.Execute
       → order.NewOrder(seatNo)
       → order.ParseTopping("味玉")     ← 値オブジェクトのバリデーション
       → order.NewCookingNote("かため") ← 値オブジェクトのバリデーション
       → o.AddItem(item)
       → orderWriter.Save(o)

2. 食券が厨房に届く（注文確定）
   POST /orders/{id}/confirm → OrderHandler.HandleConfirm
     → ConfirmOrderUsecase.Execute
       → orderReader.FindByID(id)
       → o.Confirm(menuReader)
         ├── 【業務ルール】品切れチェック → menu.IsAvailable()
         ├── 【業務ルール】トッピング3つまで
         ├── 【状態変更】status = Confirmed
         └── 【ドメインイベント発行】OrderConfirmed を溜める
       → orderWriter.Save(o)
       → eventPublisher.Publish(OrderConfirmed)  ← ここで購読者に届く

3. 厨房が調理を開始する（イベント購読）
   OrderConfirmed イベントを StartCookingUsecase が購読
     → StartCookingUsecase.HandleOrderConfirmed
       → kitchenReader.FindByID(kitchenID)
       → k.AddCookingTask(orderID)
         └── 【業務ルール】同時調理10個まで
       → kitchenWriter.Save(k)

4. 調理が完了する
   POST /kitchen/complete → KitchenHandler.HandleCompleteCooking
     → CompleteCookingUsecase.Execute
       → kitchenReader.FindByID(kitchenID)
       → k.CompleteCookingTask(orderID)
         ├── 【業務ルール】すでに完了してないかチェック
         ├── 【状態変更】task.status = Completed
         └── 【ドメインイベント発行】CookingCompleted
       → kitchenWriter.Save(k)
       → eventPublisher.Publish(CookingCompleted)

5. 待ち時間を確認する（ドメインサービス経由）
   GET /wait-time → OrderHandler.HandleWaitTime
     → EstimateWaitTimeUsecase.Execute
       → WaitTimeCalculator.EstimateWaitTime(kitchenID)
         → orderReader.CountPending()        ← Order 集約のデータ
         → kitchenReader.FindByID(kitchenID) ← Kitchen 集約のデータ
         → 計算して返す（状態を変えない）

6. メニューが品切れになる
   POST /menus/{id}/soldout → MenuHandler.HandleSoldOut
     → SoldOutMenuUsecase.Execute
       → menuReader.FindByID(menuID)
       → m.SoldOut()                  ← 状態変更
       → menuWriter.Save(m)
     → 次に Order.Confirm() が呼ばれた時、品切れチェックで弾かれる
```

## 集約間の連携

```
Order 集約               Kitchen 集約
  Confirm()                AddCookingTask()
    ↓                        ↓
  OrderConfirmed ────→ StartCookingUsecase
    （イベント購読）         CompleteCookingTask()
                              ↓
                           CookingCompleted
                           CookingRejected ────→ CancelOrderUsecase
                             （補償アクション）

Order は Kitchen を知らない。イベント経由で疎結合に連携する。
```

## 4パターンの比較

| パターン | 対象 | usecase層 | 集約 | イベント | テーブル |
|---------|------|----------|------|--------|--------|
| Transaction Script | closing/ | なし | なし | なし | なし |
| Active Record | staff/ | なし | なし | なし | staffs |
| Domain Model | kitchen/, menu/ | あり | あり | あり | kitchens, cooking_tasks |
| Event Sourcing | order/ | あり | あり | あり（状態の源泉） | events + projections |

## 関連ドキュメント

- [Event Design](event-design.md) — イベントの構造、メタデータ、BC間連携
- [Event Sourcing](event-sourcing.md) — ES版Order詳細
- [Architecture](architecture.md) — レイヤー構成
