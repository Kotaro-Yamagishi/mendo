# mendo（麺道）

Go + DDD + Clean Architecture のリファレンス実装

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean-brightgreen?style=flat-square)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
[![DDD](https://img.shields.io/badge/DDD-Domain--Driven%20Design-blue?style=flat-square)](https://www.domainlanguage.com/ddd/)

ラーメンチェーン「麺道」の店舗システムを題材に、DDD の主要パターンを1つのプロジェクトで実演する学習用リポジトリ。

## 実装パターン一覧

| パターン | 実装箇所 | 説明 |
|---------|---------|------|
| Domain Model | Kitchen, Menu | 集約 + エンティティ + 値オブジェクト |
| Event Sourcing | Order | イベント列で状態を管理 |
| Active Record | Staff | 構造体 + CRUD。補完領域向け |
| Transaction Script | Closing | 関数1つ。最もシンプル |
| CQRS | Order | 書き込みと読み取りを分離 |
| Saga | Kitchen → Order | 補償アクション（CookingRejected → Cancel） |
| Process Manager | SpecialOrder | 承認フロー。状態遷移管理 |
| Outbox + DLQ | infrastructure | イベント配信保証 + リトライ失敗退避 |
| Datasource | repository ↔ datasource | DB差し替え対応。SQLとドメイン変換の分離 |
| OLAP | Analytics Projection | 業務イベントから分析モデル構築 |

## Getting Started

Docker + Go 1.24 が必要。

```bash
# MySQL 起動
docker compose up -d

# アプリ起動
go run cmd/main.go

# ビルド確認
go build ./...

# Lint
golangci-lint run ./...
```

## ディレクトリ構成

```
mendo/
├── cmd/main.go              # エントリーポイント
├── internal/
│   ├── domain/              # 業務ルール（集約、値オブジェクト、イベント）
│   ├── application/         # ユースケース（command / query）
│   ├── infrastructure/      # 技術的実装
│   │   ├── repository/      # ドメインIF実装（InMemory / datasource経由）
│   │   ├── datasource/      # SQL抽象化（DTO入出力）
│   │   │   └── mysql/       # MySQL固有の実装
│   │   ├── eventbus/        # イベント配信（Watermill）
│   │   └── mysql/           # DB接続 + トランザクション
│   ├── interface/handler/   # HTTPハンドラ
│   ├── server/              # ルーティング + イベント購読
│   ├── staff/               # Active Record（補完領域）
│   ├── closing/             # Transaction Script（補完領域）
│   └── di/                  # DI（Google Wire）
├── docs/                    # 詳細ドキュメント
├── migrations/              # DBマイグレーション
└── docker-compose.yml
```

## ドキュメント

| ドキュメント | 内容 |
|-----------|------|
| [Business Flow](docs/business-flow.md) | 注文→調理→提供の全体フロー。イベント連携図 |
| [Architecture](docs/architecture.md) | レイヤー構成、Datasourceパターン、集約追加手順 |
| [Event Design](docs/event-design.md) | 3種類のイベント、BC間連携、マイクロサービス移行 |
| [Event Sourcing](docs/event-sourcing.md) | ES版Order、Projection、CQRS |
| [DDD Concepts](docs/ddd-concepts.md) | DDD概念とコードの対応表 |
| [Schema Design](docs/schema-design.md) | 全テーブルのスキーマ設計 |
| [Analytics](docs/analytics-design.md) | 分析モデル（OLAP）、星形スキーマ |
| [Conventions](docs/conventions.md) | コーディング規約（AI向け。厳格） |
| [Chapters](docs/chapters.md) | 書籍の章↔コードの対応表 |
| [Linting](docs/linting.md) | Lint設定、DDD依存ルール、各層の責務 |
| [ADR-001](docs/adr/001-contract-in-domain.md) | contract/ の配置決定 |

## Dependency Direction

```
handler → application → domain ← infrastructure
                                    ├─ repository（DTO↔ドメイン変換）
                                    │     ↓
                                    └─ datasource（SQL抽象化）
                                          ↓
                                       datasource/mysql
```

## Roadmap

### Infrastructure
- [ ] Kafka 統合（EventBus → Kafka KRaft）
- [ ] Schema Registry（Protobuf / Avro）

### Code Quality
- [ ] エラーハンドリング設計（Domain / Application / Infrastructure Error の分離）
- [ ] 構造化ログ（slog + CorrelationID）
- [ ] バリデーション戦略
- [ ] Graceful Shutdown + ヘルスチェック
- [ ] OpenTelemetry（分散トレーシング）

### Testing
- [ ] 集約の単体テスト
- [ ] E2E テスト（Docker Compose）
- [ ] Testcontainers

### Documentation
- [ ] OpenAPI / Swagger
- [ ] イベントカタログ

## References

### 書籍
- [ドメイン駆動設計をはじめよう](https://www.oreilly.co.jp/books/9784814400737/) — Vlad Khononov 著
- [実践ドメイン駆動設計](https://www.shoeisha.co.jp/book/detail/9784798131610) — Vaughn Vernon 著

### 動画
- [イベントストーミング入門](https://speakerdeck.com/nrslib/getting-started-with-event-storming) — nrslib

### ライブラリ
- [Google Wire](https://github.com/google/wire) — DI
- [golangci-lint](https://golangci-lint.run/) — Linter
