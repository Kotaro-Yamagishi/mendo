# バリデーション設計

## 原則

各層で「何を」「なぜ」検証するかを分離する。1つのフィールドが引っかかっても止めず、全フィールドをチェックしてまとめてエラーを返す。

## 各層のバリデーション責務

### handler 層: 入力形式の検証

API として受け入れられるかの形式チェック。ドメイン知識を必要としない。

- 必須フィールドが空でないか
- 数値が明らかにおかしくないか（負の値等）
- 配列が空でないか
- 配列の各要素の必須フィールド

#### go-playground/validator を使った実装パターン

handler 内にリクエスト用 struct を定義し、`validate` タグで形式チェックを宣言する。`validateInput` 共通ヘルパーが struct タグを読んでエラーを `ValidationWithFields` に自動変換する。

```go
// handler 内にリクエスト struct を定義（validate タグ付き）
type createOrderRequest struct {
    SeatNo int                      `json:"seat_no" validate:"required,min=1"`
    Items  []createOrderItemRequest `json:"items"   validate:"required,min=1,dive"`
}

type createOrderItemRequest struct {
    MenuID   string   `json:"menu_id"  validate:"required"`
    Toppings []string `json:"toppings"`
    Hardness string   `json:"hardness"`
}

// handler 内で readJSON → validateInput → application input に変換
func (h *OrderHandler) HandleCreate(w http.ResponseWriter, r *http.Request) error {
    var req createOrderRequest
    if err := readJSON(r, &req); err != nil {
        return err
    }
    if err := validateInput(req); err != nil {
        return err
    }
    // req → ordercommand.CreateOrderInput に変換して usecase 呼び出し
    input := ordercommand.CreateOrderInput{
        SeatNo: req.SeatNo,
        Items:  toOrderItems(req.Items),
    }
    return h.usecase.Execute(r.Context(), input)
}
```

#### なぜ handler 内にリクエスト struct を定義するか

application 層の struct（`CreateOrderInput` 等）に `validate` タグを付けると、application 層が HTTP のバリデーションルールという知識を持ってしまう。application 層は HTTP を知らないはずなので、層の分離が崩れる。

handler 内に専用の struct を定義して変換することで、各層の責務を維持する。

```
HTTP リクエスト
  ↓ readJSON
createOrderRequest   ← handler 内に定義。validate タグはここだけ
  ↓ validateInput
  ↓ 変換（toOrderItems 等）
CreateOrderInput     ← application 層。validate タグなし。HTTP を知らない
  ↓ usecase.Execute
```

#### validate タグに書いていいもの/書いてはいけないもの

**OK（形式チェック）**:
- `required` — フィールドが空でないか
- `min=1` — 数値が1以上か、配列に1個以上あるか
- `dive` — スライスの各要素に続くタグを適用する

**NG（ドメイン制約）**:
- `max=100` — 座席上限はドメインルール。`NewSeatNumber` の責務
- `oneof=かため ふつう やわらかめ` — 硬さの種類はドメインルール。`NewHardness` の責務

形式チェック（空でないか、配列に要素があるか）は handler で弾く。値の意味的な妥当性（100席以下か、有効な硬さか）は値オブジェクトが判定する。

レスポンス:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "入力内容に誤りがあります",
    "fields": [
      {"field": "seat_no", "message": "1以上の値を指定してください"},
      {"field": "items[0].menu_id", "message": "メニューIDは必須です"}
    ]
  }
}
```

ルール:
- `readJSON` の直後に `validateInput` を呼ぶ
- リクエスト struct は handler ファイル内に private struct として定義する
- struct タグには形式チェックのみ書く。ドメイン制約は書かない
- `validateInput` が全フィールドをまとめてチェックし `ValidationWithFields` で返す
- 1つ引っかかったら即 return しない

### 値オブジェクト: ドメインの制約

業務知識に基づく制約。「その値がドメインとして意味があるか」を検証する。

```go
func NewSeatNumber(n int) (SeatNumber, error) {
    if n < 1 || n > 100 {
        return 0, apperrors.Domain(ErrCodeInvalidSeatNumber, "座席番号は1〜100の範囲")
    }
    return SeatNumber(n), nil
}

func NewHardness(s string) (Hardness, error) {
    switch Hardness(s) {
    case HardnessKatame, HardnessFutsuu, HardnessYawarakame:
        return Hardness(s), nil
    default:
        return "", apperrors.Domain(ErrCodeInvalidHardness, "...")
    }
}
```

ルール:
- 全ての値オブジェクトに `New*` コンストラクタを用意する
- コンストラクタ内でバリデーション
- `apperrors.Domain` で業務エラーとして返す（Validation ではない）
- 型エイリアス（`type XxxID string`）は ID 生成が目的なのでバリデーション不要

### 集約: 状態に依存するルール

「今の状態でその操作ができるか」を検証する。

```go
func (o *Order) Confirm() error {
    if o.status != StatusPending {
        return apperrors.Domain(ErrCodeNotPending, "確定待ち以外は確定できません")
    }
    if len(o.items) == 0 {
        return apperrors.Domain(ErrCodeEmptyItems, "注文が空です")
    }
}

func (o *Order) Cancel(reason string) error {
    if reason == "" {
        return apperrors.Domain(ErrCodeEmptyReason, "キャンセル理由は必須です")
    }
    if o.status != StatusConfirmed {
        return apperrors.Domain(ErrCodeNotConfirmed, "確定済みのみキャンセルできます")
    }
}
```

ルール:
- 引数の妥当性チェック → 状態遷移チェックの順
- `apperrors.Domain` で返す
- 外部サービスの呼び出しは避ける

### application 層: バリデーションしない

オーケストレーションのみ。`return err` するだけ。

## リスト入力のパターン

### パターン A: 全件 OK か全件 NG（トランザクション的）

注文作成。1個でもダメなら全部ダメ。handler で `ValidationWithFields` を使い、配列要素はインデックス付きフィールド名で返す。

```json
{"field": "items[0].menu_id", "message": "メニューIDは必須です"}
```

### パターン B: 個別に成功/失敗（バッチ的）

メニューインポート。1件失敗しても他は続行。Job に `RecordFailure` で記録し、完了後にサマリーを返す。

## handler と値オブジェクトの検証の重複

handler で「空でないか」、値オブジェクトで「ドメインとして妥当か」を検証するため、一部重複する。これは意図的。

- handler の検証: API の契約。ドメインに到達する前にゴミを弾く
- 値オブジェクトの検証: ドメインの不変条件。どこから呼ばれても安全

handler を通さないケース（イベント経由、テスト等）でも値オブジェクトが守るため、二重チェックは必要。
