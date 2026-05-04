# Linting & Code Quality

[golangci-lint](https://golangci-lint.run/) で コード品質と DDD の依存方向を自動チェックする。

## 実行

```bash
golangci-lint run ./...
```

## コード品質チェック

| チェック項目 | ツール |
|-----------|------|
| エラー処理漏れ | errcheck, errorlint, nilerr |
| セキュリティ | gosec |
| 複雑度上限（≤ 20） | gocyclo, gocognit |
| 関数長上限（≤ 80行） | funlen |
| 未使用コード | unused, ineffassign, unparam |

設定: `.golangci.yml` — [maratori/golangci-lint-config](https://github.com/maratori/golangci-lint-config) ベースに DDD の depguard ルールを追加。

## Dependency Direction

```
interface/handler → application → domain ← infrastructure
                                   ↑           ├─ repository（DTO↔ドメイン変換）
                             全部ここに依存      │     ↓ 依存
                                               └─ datasource（SQL抽象化）
                                                     ↓ 依存
                                                  datasource/mysql（MySQL固有）
```

## DDD 依存ルール（depguard）

```
許可される依存:
  handler → application → domain ← infrastructure
                                      └─ repository → datasource IF → datasource/mysql

違反として検知:
  domain → infrastructure       "domain 層は infrastructure に依存してはいけない"
  domain → application          "domain 層は application に依存してはいけない"
  domain → database/sql         "domain 層は DB パッケージを直接使わない"
  domain → net/http             "domain 層は HTTP パッケージを直接使わない"
  application → infrastructure  "application 層は infrastructure に直接依存してはいけない"
  application → net/http        "application 層は HTTP パッケージを直接使わない"
  repository → database/sql     "repository は SQL を知らない"
  repository → datasource/mysql "repository は datasource IF に依存する。MySQL 実装に直接依存しない"
  datasource → domain/*         "datasource はドメインモデルを知らない。DTO のみ扱う"
```

## 各層の責務

| 層 | 責務 | やること | やらないこと |
|---|------|---------|------------|
| **handler** | 外部との入出力変換 | HTTP パース → usecase 呼出 → レスポンス | 業務ルール、DB、イベント発行 |
| **application** | 業務操作のオーケストレーション | ロード → コマンド → 保存 → Publish | 業務ルール判定、SQL、他 usecase の直接呼出 |
| **domain/集約** | 業務ルールと状態管理 | ルールチェック → 状態変更 → イベント append | DB、HTTP、Publish、他集約の操作 |
| **domain/service** | 集約横断の計算 | 複数集約のデータで計算 → 結果を返す | 状態保持、状態変更、イベント発行 |
| **repository** | DTO ↔ ドメインモデル変換 | datasource 経由でデータ取得 → ドメインモデルに変換 | SQL、database/sql の直接利用 |
| **datasource** | SQL 操作の抽象化 | DTO を入出力として SQL 実行 | ドメインモデルの知識 |
| **datasource/mysql** | MySQL 固有の実装 | MySQL の SQL を書く | ドメインモデルの知識 |
| **server** | サーバー設定 | ルーティング、イベント購読登録 | 業務ルール、DI |
| **di** | 依存関係の組立 | Wire プロバイダ定義 | 業務ロジック |
| **cmd/main.go** | エントリーポイント | 設定 → DI → server.Run() | それ以外の全て |

## 未使用コードの確認

Go の `unused` linter はエクスポートされた型を検知しない。新しい型・関数を定義したら手動で確認する:

```bash
grep -r "関数名" --include="*.go"
```

定義ファイル以外から参照されていなければ、未使用の可能性がある。

## 関連ドキュメント

- [Conventions](conventions.md) — コーディング規約（AI向け）
- [Architecture](architecture.md) — レイヤー構成詳細
