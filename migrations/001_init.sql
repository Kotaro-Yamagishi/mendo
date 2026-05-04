-- ============================================================
-- mendo マイグレーション
-- BC ごとにテーブル名にプレフィックスを付けて論理分離。
-- マイクロサービス化時はプレフィックスなしで別 DB に移行する。
-- ============================================================

-- ============================================================
-- 共通テーブル（全BCで同じ構造）
-- ============================================================

-- EventStore（イベントソーシング）
-- Order 集約が使用。append-only。
CREATE TABLE IF NOT EXISTS events (
    event_id       VARCHAR(255) NOT NULL,
    aggregate_id   VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type     VARCHAR(100) NOT NULL,
    version        INT          NOT NULL,
    payload        JSON         NOT NULL,
    correlation_id VARCHAR(255) NOT NULL DEFAULT '',
    causation_id   VARCHAR(255) NOT NULL DEFAULT '',
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id),
    UNIQUE KEY uq_aggregate_version (aggregate_id, version),
    INDEX idx_aggregate_id (aggregate_id),
    INDEX idx_event_type (event_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Snapshots（集約の復元高速化。将来用）
CREATE TABLE IF NOT EXISTS snapshots (
    aggregate_id   VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    version        INT          NOT NULL,
    state          JSON         NOT NULL,
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (aggregate_id, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Outbox（イベント配信保証）
CREATE TABLE IF NOT EXISTS outbox (
    id             VARCHAR(255) NOT NULL,
    event_type     VARCHAR(100) NOT NULL,
    aggregate_id   VARCHAR(255) NOT NULL DEFAULT '',
    payload        JSON         NOT NULL,
    delivered      BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_undelivered (delivered, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Dead Letter Queue（リトライ失敗イベント）
CREATE TABLE IF NOT EXISTS dlq (
    id             VARCHAR(255) NOT NULL,
    event_type     VARCHAR(100) NOT NULL,
    payload        JSON         NOT NULL,
    error          TEXT         NOT NULL,
    fail_count     INT          NOT NULL DEFAULT 0,
    handler_name   VARCHAR(255) NOT NULL DEFAULT '',
    last_fail_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 注文コンテキスト（order_context）
-- ============================================================

-- Order の Projection（リードモデル）
CREATE TABLE IF NOT EXISTS oc_order_projections (
    order_id       VARCHAR(255) NOT NULL,
    seat_no        INT          NOT NULL DEFAULT 0,
    items          JSON         NOT NULL,
    total          INT          NOT NULL DEFAULT 0,
    status         VARCHAR(50)  NOT NULL DEFAULT 'pending',
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (order_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Menu（メニューマスタ。ドメインモデルパターン）
CREATE TABLE IF NOT EXISTS oc_menus (
    menu_id        VARCHAR(255) NOT NULL,
    name           VARCHAR(255) NOT NULL,
    price          INT          NOT NULL DEFAULT 0,
    available      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 厨房コンテキスト（kitchen_context）
-- ============================================================

-- Kitchen（厨房。集約ルート）
CREATE TABLE IF NOT EXISTS kc_kitchens (
    kitchen_id     VARCHAR(255) NOT NULL,
    max_capacity   INT          NOT NULL DEFAULT 5,
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (kitchen_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- CookingTask（内部エンティティ。Kitchen に従属）
CREATE TABLE IF NOT EXISTS kc_cooking_tasks (
    task_id        VARCHAR(255) NOT NULL,
    kitchen_id     VARCHAR(255) NOT NULL,
    order_id       VARCHAR(255) NOT NULL,
    status         VARCHAR(50)  NOT NULL DEFAULT 'cooking',
    instructions   JSON         NOT NULL,
    started_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at   TIMESTAMP    NULL,
    PRIMARY KEY (task_id),
    INDEX idx_kitchen_id (kitchen_id),
    INDEX idx_order_id (order_id),
    CONSTRAINT fk_cooking_task_kitchen
        FOREIGN KEY (kitchen_id) REFERENCES kc_kitchens(kitchen_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- OrderBoard（横断 Projection。Order + Kitchen のイベントから構築）
CREATE TABLE IF NOT EXISTS kc_order_board (
    order_id       VARCHAR(255) NOT NULL,
    seat_no        INT          NOT NULL DEFAULT 0,
    order_status   VARCHAR(50)  NOT NULL DEFAULT '',
    cooking_status VARCHAR(50)  NOT NULL DEFAULT '',
    ordered_at     TIMESTAMP    NULL,
    cooking_at     TIMESTAMP    NULL,
    PRIMARY KEY (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 特別注文コンテキスト（specialorder_context）
-- ============================================================

CREATE TABLE IF NOT EXISTS sc_special_orders (
    id             VARCHAR(255) NOT NULL,
    order_id       VARCHAR(255) NOT NULL,
    menu_name      VARCHAR(255) NOT NULL,
    status         VARCHAR(50)  NOT NULL DEFAULT 'pending',
    reject_reason  VARCHAR(255) NOT NULL DEFAULT '',
    suggested_menu VARCHAR(255) NOT NULL DEFAULT '',
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- 補完領域
-- ============================================================

-- Staff（アクティブレコード）
CREATE TABLE IF NOT EXISTS staffs (
    id             VARCHAR(255) NOT NULL,
    name           VARCHAR(255) NOT NULL,
    phone          VARCHAR(50)  NOT NULL DEFAULT '',
    shift_type     VARCHAR(50)  NOT NULL DEFAULT '',
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
