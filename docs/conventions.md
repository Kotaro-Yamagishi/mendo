# コーディング規約

AI にコードを書かせる前提で、厳格なルールを定義する。次のプロジェクトでも使い回す。

## 命名規則

### ファイル名

| 対象 | 規則 | 例 |
|------|------|----|
| 集約ルート | `{集約名}.go` | kitchen.go |
| 値オブジェクト | `values.go` | values.go |
| イベント | `events.go`（複数形） | events.go |
| リポジトリ IF | `repository.go` | repository.go |
| InMemory 実装 | `{対象}_inmemory.go` | kitchen_inmemory.go |
| Datasource 実装 | `{対象}_ds.go` | kitchen_ds.go |
| Repository 実装 | `{対象}_repo.go` | kitchen_repo.go |
| ハンドラ | `{集約名}_handler.go` | kitchen_handler.go |
| ユースケース | `{操作名}.go` | confirm.go, cancel.go |

### 型名

| 対象 | 規則 | 例 |
|------|------|----|
| 集約ルート | `{集約名}` | Kitchen, Order |
| 値オブジェクト | 意味のある名前 | OrderID, SeatNumber, Price |
| イベント | 過去形 | OrderConfirmed, CookingCompleted |
| イベント定数 | `EventType{イベント名}` | EventTypeOrderConfirmed |
| リポジトリ IF | `Reader` / `Writer` | Reader, Writer |
| DTO | `{対象}Row` | KitchenRow, MenuRow |
| datasource IF | `{対象}DataSource` | KitchenDataSource |

### メソッド名

| 目的 | 規則 | 例 |
|------|------|----|
| コマンド | 動詞 | Confirm, Cancel, AddItem |
| クエリ | Find / Count / List | FindByID, CountPending |
| 復元 | `Reconstruct{型名}` | ReconstructKitchen |
| 新規作成 | `New{型名}` | NewKitchen |

## 禁止事項

- [ ] domain 層から infrastructure を import しない
- [ ] application 層から infrastructure を import しない
- [ ] repository から `database/sql` を import しない
- [ ] repository から `datasource/mysql` を直接 import しない
- [ ] datasource からドメイン型を import しない
- [ ] handler で `http.Error()` を使わない（`writeError()` を使う）
- [ ] DB 復元に `New*` を使わない（`Reconstruct*` を使う）
- [ ] イベント名をハードコーディングしない（定数を使う）
- [ ] 集約の外から集約のフィールドを直接変更しない（メソッド経由）

## 依存の方向性

```
handler → application → domain ← infrastructure
                                    ├─ repository（DTO↔ドメイン変換）
                                    │     ↓ 依存
                                    └─ datasource（SQL実行）
                                          ↓ 依存
                                       datasource/mysql（MySQL固有）
```

矢印の逆方向の import は depguard で禁止。

## レスポンス形式

```go
// 成功
writeSuccess(w, http.StatusOK, data)
// → {"data": {...}}

// エラー
writeError(w, http.StatusBadRequest, "invalid request")
// → {"error": "invalid request"}
```

## エラーハンドリング

### 必須ルール
- [ ] domain 層のエラーは `apperrors.Domain(code, message)` を使う
- [ ] errors.New / fmt.Errorf を domain 層で使わない
- [ ] application 層で domain エラーをラップしない（そのまま返す）
- [ ] infrastructure エラーは `apperrors.Infrastructure(message, cause)` でラップ
- [ ] handler は `AppHandlerFunc`（error を返す）で定義
- [ ] 全エンドポイントを `ErrorMiddleware` でラップ
- [ ] http.Error() を使わない
- [ ] handler でステータスコードを直接指定しない
- [ ] エラーコード定数は各集約の `errors.go` に定義
- [ ] エラーは1回だけ handle する（ログ OR 復旧 OR 表示）
- [ ] サーバーエラー（500系）はメッセージを隠蔽

## テスト方針

| パターン | テスト方針 |
|--------|---------|
| DM / ES（中核） | ピラミッド型。集約の単体テスト重視 |
| AR（補完） | ダイヤモンド型。統合テスト重視 |
| TS（補完） | 逆ピラミッド型。E2E 重視 |
