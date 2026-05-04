-- ============================================================
-- 分析系モデル（OLAP）
-- データメッシュ: 注文BCの分析データプロダクト
-- CQRS の Projection として、業務イベントから構築する
-- ============================================================

-- ============================================================
-- 特性テーブル（Dimensions）
-- ============================================================

-- 日付特性（全分析モデル共通）
CREATE TABLE IF NOT EXISTS dim_date (
    date_key     INT          NOT NULL,
    full_date    DATE         NOT NULL,
    year         INT          NOT NULL,
    month        INT          NOT NULL,
    day          INT          NOT NULL,
    day_of_week  VARCHAR(10)  NOT NULL,
    is_weekend   BOOLEAN      NOT NULL DEFAULT FALSE,
    is_holiday   BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (date_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- メニュー特性
CREATE TABLE IF NOT EXISTS dim_menu (
    menu_key     VARCHAR(255) NOT NULL,
    name         VARCHAR(255) NOT NULL,
    category     VARCHAR(100) NOT NULL DEFAULT '',
    price        INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (menu_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 座席特性
CREATE TABLE IF NOT EXISTS dim_seat (
    seat_key     INT          NOT NULL,
    seat_number  INT          NOT NULL,
    area         VARCHAR(50)  NOT NULL DEFAULT '',
    PRIMARY KEY (seat_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 事実テーブル（Facts）
-- ============================================================

-- 注文事実（OrderConfirmed イベントから生成）
-- 1注文に複数商品 → 商品ごとに1行
CREATE TABLE IF NOT EXISTS fact_orders (
    fact_id      VARCHAR(255) NOT NULL,
    order_id     VARCHAR(255) NOT NULL,
    date_key     INT          NOT NULL,
    menu_key     VARCHAR(255) NOT NULL,
    seat_key     INT          NOT NULL,
    amount       INT          NOT NULL DEFAULT 0,
    quantity     INT          NOT NULL DEFAULT 1,
    ordered_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (fact_id),
    INDEX idx_order_id (order_id),
    INDEX idx_date_key (date_key),
    INDEX idx_menu_key (menu_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 調理事実（CookingCompleted イベントから生成）
CREATE TABLE IF NOT EXISTS fact_cooking (
    fact_id      VARCHAR(255) NOT NULL,
    order_id     VARCHAR(255) NOT NULL,
    date_key     INT          NOT NULL,
    menu_key     VARCHAR(255) NOT NULL,
    started_at   TIMESTAMP    NULL,
    completed_at TIMESTAMP    NULL,
    duration_sec INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (fact_id),
    INDEX idx_order_id (order_id),
    INDEX idx_date_key (date_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 初期データ: dim_date（2026年分）
-- ============================================================
-- 本番では日付マスタ生成スクリプトで投入する。
-- 学習用に数日分だけ INSERT する。

INSERT IGNORE INTO dim_date (date_key, full_date, year, month, day, day_of_week, is_weekend, is_holiday) VALUES
(20260427, '2026-04-27', 2026, 4, 27, '月', FALSE, FALSE),
(20260428, '2026-04-28', 2026, 4, 28, '火', FALSE, FALSE),
(20260429, '2026-04-29', 2026, 4, 29, '水', FALSE, FALSE),
(20260430, '2026-04-30', 2026, 4, 30, '木', FALSE, FALSE);
