-- ============================================================
-- Seed Data（開発・テスト用初期データ）
-- 本番環境では実行しない
-- ============================================================

-- 厨房マスター
INSERT IGNORE INTO kc_kitchens (kitchen_id, max_capacity, created_at) VALUES
('kitchen-1', 10, NOW());

-- メニューマスター
INSERT IGNORE INTO oc_menus (menu_id, name, price, available, created_at, updated_at) VALUES
('menu-ramen-001', '醤油ラーメン', 800, TRUE, NOW(), NOW()),
('menu-ramen-002', '味噌ラーメン', 900, TRUE, NOW(), NOW()),
('menu-ramen-003', '塩ラーメン', 850, TRUE, NOW(), NOW()),
('menu-tsukemen-001', 'つけ麺', 950, TRUE, NOW(), NOW());
