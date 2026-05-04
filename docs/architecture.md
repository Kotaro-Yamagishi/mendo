# mendo アーキテクチャ

## このアプリは何か

mendo（麺道）は、ラーメンチェーンの注文管理システム。DDD + Clean Architecture の学習用リファレンス実装。

以下の DDD パターンを1つのプロジェクトで実演する:
- 4つの実装パターン（Transaction Script, Active Record, Domain Model, Event Sourcing）
- CQRS + Projection
- Outbox Pattern + DLQ
- サーガ（補償アクション）+ プロセスマネージャー
- BC 間連携（公開イベント / contract）
- 分析モデル（OLAP / データメッシュ）

## アーキテクチャの全体像

```
interface（handler）
  ↓ 依存
application（usecase）
  ↓ 依存
domain（集約、値オブジェクト、イベント、リポジトリ IF）
  ↑ 実装
infrastructure
  ├─ repository/（ドメイン IF の実装。DTO ↔ ドメインモデル変換）
  │   ├─ *_inmemory.go（InMemory 実装）
  │   └─ *_repo.go（datasource を使う実装）
  ├─ datasource/（SQL 操作の抽象化。DTO 入出力）
  │   ├─ dto.go（全 DTO 定義）
  │   ├─ *.go（datasource IF 定義）
  │   └─ mysql/（MySQL 実装）
  ├─ mysql/（DB 接続 + トランザクション管理）
  ├─ eventbus/（イベント配信）
  ├─ outbox/（リレーサービス）
  └─ analytics/（分析モデル投影）
```

## レイヤーの責務

| レイヤー | 責務 | 知っていいこと | 知ってはいけないこと |
|--------|------|-----------|-------------|
| domain | 業務ルール | 自分自身のみ | DB, HTTP, 他の集約の内部 |
| application | ユースケースの orchestration | domain IF | DB, HTTP |
| interface | HTTP リクエスト/レスポンス | application | DB, infrastructure |
| repository | DTO ↔ ドメインモデル変換 | domain, datasource IF | SQL, 具体的な DB |
| datasource | SQL 実行 | DTO のみ | ドメインモデル |
| datasource/mysql | MySQL 固有の SQL | DTO, SQL | ドメインモデル |

## Datasource パターン

### なぜ分離するか

```
Before（repository に SQL が直接ある）:
  MySQL → PostgreSQL に変えたら全 repository を書き直し
  repository の中で SQL とドメイン変換が混在

After（datasource 層を分離）:
  MySQL → PostgreSQL なら datasource/mysql/ → datasource/postgres/ に差し替え
  repository はそのまま（DTO ↔ ドメインモデル変換は DB に依存しない）
```

### Decorator パターン

datasource IF を満たすラッパーで横断的関心事を挟める:
- リトライ（指数バックオフ）
- クエリログ / トレーシング
- メトリクス
- キャッシュ
- サーキットブレーカー

### InMemory / MySQL の切り替え

DI（di/providers.go）で切り替える:
- InMemory: repository の `*_inmemory.go` を使う
- MySQL: `datasource/mysql/` → repository の `*_repo.go` を使う

## 新しい集約を追加する手順

1. `domain/{集約名}/` に集約ルート、値オブジェクト、イベント、リポジトリ IF を作成
2. `domain/{集約名}/repository.go` に Reader / Writer IF を定義
3. `infrastructure/datasource/dto.go` に DTO を追加
4. `infrastructure/datasource/{集約名}.go` に datasource IF を定義
5. `infrastructure/datasource/mysql/{集約名}_ds.go` に MySQL 実装
6. `infrastructure/repository/{集約名}_inmemory.go` に InMemory 実装
7. `infrastructure/repository/{集約名}_repo.go` に datasource を使う実装
8. `application/command/{集約名}/` にユースケースを作成
9. `interface/handler/{集約名}_handler.go` にハンドラを作成
10. `di/providers.go` に DI 配線を追加
11. `go build ./...` と `golangci-lint run ./...` で確認
