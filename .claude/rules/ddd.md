# DDD + Clean Architecture コーディングルール

## 依存の方向性

全レイヤーが domain に向かって依存する。domain は何にも依存しない。

```
interface/handler → application → domain ← infrastructure
                                   ↑
                             全部ここに依存
```

- domain 層は外部パッケージ（DB, HTTP, メッセージキュー等）を import しない
- application 層は infrastructure 層を直接参照しない。domain 層の IF 経由で使う
- infrastructure 層は domain 層の IF を実装する

## 各層の責務

### interface/handler
- **責務**: 外部との入出力の変換
- やること: HTTP リクエストのパース → ユースケース呼び出し → レスポンス返却
- やらないこと: 業務ルールの判定、DB アクセス、イベント発行
- **1ハンドラ = 1ユースケース**。handler で複数のユースケースを呼ばない。後続処理はイベント経由で連携する
- 複数の操作を1アクションでやりたい場合は、ユースケース側で統合する（`CreateAndConfirmOrderUsecase` 等）

### application（ユースケース）
- **責務**: 業務操作のオーケストレーション
- やること: 集約のロード → コマンド実行 → 保存 → イベント配信（Publish）
- やらないこと: 業務ルールの判定（集約の仕事）、SQL を書く（infra の仕事）、HTTP の処理（handler の仕事）
- **他のユースケースを直接呼ばない**。イベント経由で連携する
- 1ユースケース = 1ファイル = 1つの業務操作

### domain/集約
- **責務**: 業務ルールの実行と状態管理
- やること: 業務ルールのチェック → 状態変更 → ドメインイベントを溜める（append）
- やらないこと: DB アクセス、HTTP 通信、イベントの Publish（溜めるだけ。配信はユースケースの責務）、他の集約の直接操作
- 集約のルートが外部からの唯一のアクセス入口

### domain/service
- **責務**: 集約をまたぐステートレスな計算
- やること: 複数の集約やリポジトリのデータを使って計算し、結果を返す
- やらないこと: 状態を持つ、状態を変更する、イベントを発行する

### infrastructure
- **責務**: 技術的な実装の詳細
- やること: domain の IF を実装する（Repository, EventStore, EventPublisher）、SQL を書く、外部 API と通信する
- やらないこと: 業務ルールの判定、ユースケースの呼び出し

### server
- **責務**: HTTP サーバーの起動設定とイベント購読の登録
- やること: ルーティング定義（routes.go）、イベント購読者の登録（subscribers_*.go）、サーバー設定
- やらないこと: 業務ルールの判定、DI

### di
- **責務**: 全レイヤーの依存関係の組み立て
- やること: Wire でプロバイダを定義し、具体実装を IF に注入する
- やらないこと: 業務ロジック、ルーティング

### cmd/main.go
- **責務**: エントリーポイント
- やること: 設定値の定義、DI でアプリ組み立て、server.Run() を呼ぶ
- やらないこと: 上記以外の全て

## 集約のルール

### 構成
- 集約ごとにパッケージを分ける（`domain/{集約名}/`）
- 1集約のファイル構成は統一する:

| ファイル | 役割 | 必須 |
|---------|------|------|
| `{集約名}.go` | 集約のルート。コマンド + 業務ルール | 必須 |
| `values.go` | 値オブジェクト全て | 必須 |
| `events.go` | ドメインイベント定義 | 必須 |
| `repository.go` | 全 IF（Reader, Writer, ProjectionReader, ProjectionWriter 等）を1ファイルにまとめる | 必須 |
| `projection.go` | リードモデル（ES の場合） | 任意 |
| `{内部エンティティ名}.go` | 内部エンティティ | 任意 |

### 業務ルール
- 業務ルールは集約のルートのコマンドメソッド内に書く
- application 層や handler 層に業務ルールを書かない
- 業務ルールは集約の状態だけで判定する。外部サービスの呼び出しは避ける

### 集約のルート
- 外部からの唯一のアクセス入口
- 内部エンティティを直接操作させない。ルート経由でのみ変更
- フィールドは private（小文字）。アクセサ経由で公開

### 値オブジェクト
- 不変（Immutable）。作成後に変更しない
- バリデーションはコンストラクタ（New*, Parse*）に内蔵
- string や int をそのまま使わず、意味のある型にする（`UserID`, `Price`, `Priority`）

### ドメインイベント
- 過去形で命名する（`OrderConfirmed`, `CookingCompleted`）
- `domain.DomainEvent` を埋め込む（EventID, CorrelationID 等のメタデータ）
- 集約内では `append` で溜めるだけ。Publish はユースケースの責務
- イベントにビジネスロジックを載せない。データの入れ物

- イベント名（EventType）と集約種別（AggregateType）は **必ず定数で定義** する。文字列リテラルを直接使わない
  ```go
  // ✅ 正しい
  const AggregateTypeOrder = "Order"
  const EventTypeOrderConfirmed = "order.confirmed"
  domain.NewDomainEvent(id, AggregateTypeOrder, EventTypeOrderConfirmed, corr)

  // ❌ ダメ
  domain.NewDomainEvent(id, "Order", "order.confirmed", corr)
  ```
- Subscribe 時もイベント名定数を使う: `bus.Subscribe(order.EventTypeOrderConfirmed, ...)`

### ドメインサービス
- `domain/service/` に配置。特定の集約に属さない
- ステートレス（自分は状態を持たない）
- 複数の集約やリポジトリのデータを使う計算に限定

## application 層のルール

詳細は上記「各層の責務」を参照。以下は追加の細則:

- 1ユースケース = 1ファイル = 1つの業務操作
- ユースケースは薄いオーケストレーション: ロード → コマンド実行 → 保存 → イベント配信
- ユースケースに業務ルールを書かない。集約に委譲する
- IF を事前定義しない。struct を直接使う

## infrastructure 層のルール

詳細は上記「各層の責務」を参照。以下は追加の細則:

- リポジトリ実装は `infrastructure/repository/` に配置
- イベントバス実装は `infrastructure/eventbus/` に配置
- イベントストア実装は `infrastructure/eventstore/` に配置
- 横断リードモデル（複数集約のイベントを使う Projection）は `infrastructure/projection/` に配置
- domain 層の IF を実装する

## interface 層のルール

詳細は上記「各層の責務」を参照。以下は追加の細則:

- HTTP ハンドラは `interface/handler/` に配置
- ユースケースを呼ぶだけ。業務ルールを書かない
- リクエストのパース → ユースケース呼び出し → レスポンス返却

## DI のルール

詳細は上記「各層の責務」を参照。以下は追加の細則:

- `di/` で Wire を使って全レイヤーを配線
- `cmd/main.go` で DI 結果を受け取り、イベント購読を登録し、ルーティングを設定
- main.go と di/ だけが全レイヤーを知っている

## リードモデル（Projection）の配置

```
1つの集約のイベントだけ使う → domain/{集約名}/projection.go
複数の集約のイベントを使う → infrastructure/projection/
```

## イベント購読（subscribers）のルール

- イベント購読の登録は `server/subscribers.go`（または `server/subscribers_{集約名}.go`）に配置
- `cmd/main.go` にイベント購読を書かない
- 購読者が増えたら集約ごとにファイル分割する
- 購読者の中ではユースケースのメソッドを呼ぶだけ。業務ルールを書かない

## 型エイリアス（ID, 値オブジェクト）の配置

- その型を「所有する」集約の `values.go` に定義する
- 他の集約は import して使う（`user.UserID`）
- どの集約にも属さない汎用型は `domain/shared/`（Money, Timestamp 等）

## BC 間の公開イベント（contract/）

- 公開イベントの型は `internal/domain/contract/` に定義する
- BC 間のデータ交換は contract/ の型のみ使う
- 他の BC の `domain/` を直接 import しない
- 変換関数（`ToPublicXxx`）は発行元の BC の `domain/` に置く（内部イベントを知っているのは発行元だけ）
- `contract/` に IF やビジネスロジックは置かない
- マイクロサービス化する際は contract/ を別リポジトリに切り出す

## Datasource 層

- DB アクセスは全て datasource IF 経由で行う
- datasource はドメイン型を import しない。DTO のプリミティブ型のみ扱う
- repository は SQL を書かない。database/sql を import しない
- datasource のメソッドはテーブル単位の細かいメソッドにする（使い回し可能）
- InMemory 実装は datasource を使わない。ドメインモデルを直接メモリに保持

## DTO

- `infrastructure/datasource/dto.go` に定義する
- プリミティブ型のみ（string, int, bool, time.Time）。ドメイン型を使わない
- 業務ロジックもバリデーションも持たない
- DB のカラムと1:1対応

## InMemory 実装

- 全て `infrastructure/repository/` に置く
- ファイル名: `{対象}_inmemory.go`

## Reconstruct ファクトリ

- DB から復元する時は `Reconstruct*` を使う
- `New*` は新規作成時のみ（ドメインイベント発行あり）
- 例: `ReconstructKitchen()`, `ReconstructMenu()`, `ReconstructSpecialOrder()`

## Handler レスポンス

- 成功: `writeSuccess(w, status, data)` で JSON
- エラー: `writeError(w, status, msg)` で JSON
- `http.Error()` は使用禁止

## ビルド・品質チェック

- 変更後は `go build ./...` と `golangci-lint run ./...` でエラー0を確認
- **未使用コードを残さない**: 新しい型・関数・メソッドを定義したら、必ず呼び出し元も実装する。「定義だけして後で使う」は禁止。Go の `unused` linter はエクスポートされた型を検知しないため、自分で確認する
- 新しいファイルを作成したら、そのファイル内の全 public 関数・型が他のファイルから参照されているか確認する
- 確認方法: `grep -r "関数名" --include="*.go"` で定義ファイル以外から参照されているか確認
