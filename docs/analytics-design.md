# 分析系モデル設計（データメッシュ）

## 概要

mendo の分析系モデル（OLAP）。業務イベントから CQRS の Projection として構築する。

データメッシュの考え方に基づき、注文BC のチームが業務系と分析系の両方を管理する。

## 星形スキーマ

```
          dim_menu
            ↑
dim_date <- fact_orders -> dim_seat
```

## テーブル

### 事実テーブル

**fact_orders**: OrderConfirmed イベントから生成。1商品 = 1行。

| カラム | 説明 |
|-------|------|
| fact_id | 一意ID |
| order_id | 注文ID |
| date_key | FK -> dim_date |
| menu_key | FK -> dim_menu |
| seat_key | FK -> dim_seat |
| amount | 金額（price × quantity） |
| quantity | 数量 |
| ordered_at | 注文確定日時 |

**fact_cooking**: CookingCompleted イベントから生成。

| カラム | 説明 |
|-------|------|
| fact_id | 一意ID |
| order_id | 注文ID |
| date_key | FK -> dim_date |
| menu_key | FK -> dim_menu |
| started_at | 調理開始日時 |
| completed_at | 調理完了日時 |
| duration_sec | 調理時間（秒） |

### 特性テーブル

**dim_date**: 日付マスタ。全分析モデル共通。
**dim_menu**: メニューマスタ。Menu 集約から導出。
**dim_seat**: 座席マスタ。SeatNumber 値オブジェクトから導出。

## データフロー

```
業務イベント（OLTP）
  OrderConfirmed
    ↓ subscriber（CQRS Projection）
  fact_orders に INSERT
  dim_menu を UPSERT（メニュー情報更新）

  CookingCompleted
    ↓ subscriber
  fact_cooking に INSERT
```

## 分析クエリ例

曜日別売上:
```sql
SELECT dd.day_of_week, SUM(f.amount)
FROM fact_orders f
JOIN dim_date dd ON f.date_key = dd.date_key
GROUP BY dd.day_of_week
```

メニュー別平均調理時間:
```sql
SELECT dm.name, AVG(fc.duration_sec) / 60.0 AS avg_min
FROM fact_cooking fc
JOIN dim_menu dm ON fc.menu_key = dm.menu_key
GROUP BY dm.name
```

## DDD との対応

| DDD の概念 | 分析モデルへの対応 |
|---|---|
| ドメインイベント | -> 事実テーブルの行 |
| 集約の属性 | -> 特性テーブル |
| BC の境界 | -> 分析モデルの所有権の境界 |
| 公開された言葉 | -> 分析データプロダクトのスキーマ |
| CQRS Projection | -> 事実テーブルへの投影処理 |
