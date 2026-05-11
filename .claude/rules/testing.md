# テストルール

## 共通ルール

### 構造
- テストは **Table-Driven** を基本とする。同じメソッドに対する複数ケースは1つのテーブルにまとめる
- 正常系と異常系は **別テスト関数** にする。テストロジック内に if 分岐を入れない
- テストの構造は **Given-When-Then**（または Arrange-Act-Assert）に従う
- 1テーブルのフィールド数は **4以下**。超えるなら関数を分割する

### 命名
- テスト関数名: `Test_{対象}_{状況}` の形式。日本語で可読性を優先
  - 例: `Test_Order_Confirm_ステータスチェック`, `Test_MenuRepo_FindByID`
- サブテスト名: Given の状態が読み取れる名前
  - 例: `"pending_アイテムあり_成功"`, `"confirmed_二重確定_エラー"`

### パッケージ
- **ブラックボックステスト（`package xxx_test`）がデフォルト**。非公開メソッドにアクセスしない
- ホワイトボックスは複雑な内部ロジックを単体テストする場合のみ例外的に使う

### 並列実行
- 純粋関数・Stub を使うテストには `t.Parallel()` を必ずつける
- goroutine を使うテスト（eventbus, relay 等）は並列化しない
- 統合テスト（実 DB）は並列化しない

### ヘルパー
- テストヘルパーには `t.Helper()` を必ずつける
- ヘルパーは **積み上げ式** で設計する（`newPendingWithItems → newConfirmedOrder → newCanceledOrder`）
- テストダブルは `internal/testutil/` に共通パッケージとして配置する

### アサーション
- **前提条件** → `require`（失敗で即停止）
- **検証** → `assert`（続行して全結果を見る）
- エラー検証は `require.Error` + `assert.Contains(t, err.Error(), "期待文字列")` で内容まで確認する

### 定数
- イベント型は **定数を使う**。文字列リテラルでハードコードしない
  - ✅ `order.EventTypeOrderCreated`
  - ❌ `"order.created"`
- ステータスは層による:
  - Domain/Application 層: ドメイン定数を使う（`order.StatusConfirmed`）
  - datasource/handler テスト: 文字列リテラルが正しい（DB/JSON に入る値を直接検証）

---

## 層別テストルール

### Domain 層（ユニットテスト）

**目的**: ビジネスルールが正しく実装されているか検証する

**テスト対象**:
- 集約のコマンド（状態遷移 + イベント発行）
- 値オブジェクトのバリデーション（境界値テスト）
- ReconstructFromEvents（イベント列からの状態復元）
- ドメインサービスの計算ロジック

**検証観点**:
- **状態検証**: 操作後のステータス、フィールド値
- **イベント検証**: 正しいイベントが発行されたか（型、ペイロード）
- **エラー時のイベント非発行**: 失敗した操作がイベントとして記録されていないこと

**テストダブル**: なし。Domain 層は純粋ロジック

**テストケースの導出方法**:
1. Example Mapping でルール・例・疑問を洗い出す
2. 状態遷移図から全状態×全コマンドの組み合わせを列挙する
3. 値オブジェクトの制約から境界値を特定する

```go
// イベントソーシング集約のテスト構造
func Test_Order_Confirm_ステータスチェック(t *testing.T) {
    tests := []struct {
        name    string
        setup   func(t *testing.T) *order.Order
        wantErr string
    }{...}
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            o := tt.setup(t)
            eventsBefore := len(o.UncommittedEvents())
            err := o.Confirm()
            if tt.wantErr != "" {
                require.ErrorContains(t, err, tt.wantErr)
                assert.Len(t, o.UncommittedEvents(), eventsBefore) // イベント非発行
                return
            }
            require.NoError(t, err)
            assert.Equal(t, order.StatusConfirmed, o.Status())
            // イベント検証
        })
    }
}
```

### Application 層（ユニットテスト）

**目的**: ユースケースのオーケストレーションが正しいか検証する

**テスト対象**:
- 依存先（EventStore, Outbox, Repository）が正しい引数で呼ばれるか
- エラー時に後続の処理が実行されないか
- 正しい戻り値が返るか

**検証しないこと**:
- 集約の業務ルール（Domain 層テストの責務）
- EventStore/Outbox の実装の正しさ（Infrastructure 層テストの責務）

**テストダブル**: `internal/testutil/` の Stub/Spy を使う

```go
func Test_ConfirmOrderES_正常系(t *testing.T) {
    t.Parallel()
    es := &testutil.StubEventStore{Events: []domain.Event{...}}
    outbox := &testutil.SpyOutbox{}
    pub := &testutil.SpyEventPublisher{}
    uc := apporder.NewConfirmOrderESUsecase(es, outbox, pub)

    err := uc.Execute(context.Background(), "order-1")

    require.NoError(t, err)
    require.Len(t, es.Saved, 1)        // EventStore に保存された
    require.Len(t, outbox.Stored, 1)    // Outbox に保存された
}
```

### Infrastructure/repository 層（ユニットテスト）

**目的**: ドメイン型 ↔ DTO の変換が正しいか検証する

**テスト対象**:
- Save: ドメインモデル → DTO への変換（フィールドマッピング、JSON シリアライズ）
- Load/FindByID: DTO → ドメインモデルへの変換（Reconstruct の正しさ）
- エラーハンドリング（datasource がエラーを返した場合）

**テストダブル**: `internal/testutil/` の StubXxxDataSource を使う

### Infrastructure/datasource 層（統合テスト）

**目的**: SQL が正しく動作するか検証する

**テスト対象**:
- Insert → Find のラウンドトリップ（全フィールドが保存・復元される）
- Update 後の Find（状態変更が反映される）
- 存在しない ID の Find（nil 返却、エラーにならない）
- 重複キーの Insert（PRIMARY KEY 制約）

**実行条件**:
- ビルドタグ `//go:build integration`
- Docker MySQL が必要（`make test-integration`）
- `TestMain` でマイグレーション実行（`migrations/*.sql` を読み込む）
- `t.Cleanup` でテストデータを削除

**マイグレーション**: `migrations/*.sql` を直接読み込む。テスト内にハードコードしない

### Interface/handler 層（ユニットテスト）

**目的**: HTTP リクエスト/レスポンスの変換が正しいか検証する

**テスト対象**:
- リクエストパース（不正 JSON → 400）
- ステータスコード（成功 → 200/201、業務エラー → 422）
- レスポンス JSON の構造（`{"data": {...}}` / `{"error": "..."}` ）

**検証しないこと**:
- 業務ロジック（Application/Domain 層の責務）

**テストダブル**: ユースケースの依存を Stub にして、本物のユースケースをハンドラに渡す

```go
func Test_OrderHandler_HandleConfirm(t *testing.T) {
    tests := []struct {
        name       string
        events     []domain.Event
        wantStatus int
    }{...}
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            es := &testutil.StubEventStore{Events: tt.events}
            // ... handler 構築 ...
            req := httptest.NewRequest(http.MethodPost, "/orders/order-1/confirm", nil)
            req.SetPathValue("id", "order-1")
            rec := httptest.NewRecorder()
            h.HandleConfirm(rec, req)
            assert.Equal(t, tt.wantStatus, rec.Code)
        })
    }
}
```

### Contract テスト（ユニットテスト）

**目的**: BC 間の公開イベントのスキーマが維持されているか検証する

**テスト対象**:
- スキーマ検証（必須フィールドが存在するか）
- JSON ラウンドトリップ（シリアライズ→デシリアライズでデータが壊れないか）
- 消費側との互換性（公開イベント → 消費側のドメインモデルへの変換）
- 後方互換性（フィールド追加/欠損でエラーにならないか）

**ファイル構成**: ソースファイルと 1:1 対応（`order.go` → `order_test.go`）

### E2E テスト（統合テスト）

**目的**: ビジネスフロー全体が端から端まで通るか検証する

**テスト対象**:
- 主要ビジネスフローの正常系のみ（2-5 本）
- 各ステップの HTTP ステータスコード
- 最終状態の DB 検証

**書かないもの**:
- 異常系（Unit テストで保証済み）
- バリデーションエラー（Handler テストで保証済み）

**ファイル構成**:
```
internal/e2e/
  testmain_test.go          ← DB接続・マイグレーション・seed・サーバー起動
  helper_test.go             ← HTTP送信・レスポンス解析・DB検証ヘルパー
  {フロー名}_flow_test.go    ← 1ファイル = 1ビジネスフロー
```

**マイグレーション**: `migrations/*.sql` を読み込む。seed データは `migrations/seed.sql`
**cleanup**: seed データ（kc_kitchens, oc_menus）は削除しない。トランザクションデータのみ削除

---

## テスト実行

```bash
make test              # ユニットテストのみ（高速）
make test-integration  # 統合テスト（Docker 必要）
make test-all          # 全テスト
make cover             # カバレッジをブラウザで表示
```

## テストファイルの配置

- テストファイルはソースファイルと同じディレクトリに置く
- ソースファイルと 1:1 対応させる（`xxx.go` → `xxx_test.go`）
- テストダブルは `internal/testutil/` に集約
- E2E テストは `internal/e2e/` に配置
