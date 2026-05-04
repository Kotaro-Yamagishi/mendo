# ADR-001: 公開イベント（contract）の配置場所

## ステータス
採用

## コンテキスト
BC 間で共有する公開イベントの型定義をどこに配置するか。

## 検討した選択肢

### 案1: internal/contract/（domain の外）
- depguard で domain/{他BC}/ を全面禁止できてシンプル
- ただし公開イベントもドメインの知識に基づいているのに domain の外にあるのが不自然

### 案2: internal/domain/order/contract/（各 BC 内）
- 各 BC が自分の公開 IF を管理。責務が明確
- ただし kitchen が import "domain/order/contract" するので、import パスに order が入る
- depguard で domain/order を deny すると contract も巻き込まれる

### 案3: internal/domain/contract/（domain 直下）← 採用
- 公開イベントがドメインの知識として domain/ 内に存在する
- kitchen が import "domain/contract" で、order/ の中ではない
- depguard で domain/order を deny しても domain/contract は影響なし

## 決定
案3（internal/domain/contract/）を採用。

## 理由
- 公開イベントはドメインの知識に基づいたもの。domain/ の中にいるべき
- depguard で各 BC の domain/{bc}/ を deny しつつ、domain/contract/ は許可できる
- contract パッケージは完全に自己完結（外部 import なし）。domain 層のルール（infrastructure 等の import 禁止）に抵触しない
- マイクロサービス化する場合、domain/contract/ を独立リポジトリに切り出せる

## 影響
- BC 間の公開イベントの型は domain/contract/ に定義する
- 変換関数（ToPublicXxx）は発行元の BC の domain/ に残す
- contract/ に IF やビジネスロジックは置かない
