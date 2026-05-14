-- ============================================================
-- インデックス最適化マイグレーション
-- テーマ1（インデックス戦略とクエリ最適化）の学習成果を適用
-- ============================================================

-- ============================================================
-- 改善1: 冗長インデックスの削除
-- ============================================================

-- events.idx_aggregate_id は uq_aggregate_version (aggregate_id, version) の
-- 左端プレフィックスと重複している。
-- WHERE aggregate_id = ? のクエリは uq_aggregate_version で対応可能。
-- 冗長インデックスは INSERT のたびに更新コストを払い、
-- Buffer Pool を無駄に消費するため削除する。
DROP INDEX idx_aggregate_id ON events;

-- ============================================================
-- 改善2: kc_cooking_tasks の複合インデックス最適化
-- ============================================================

-- 現状: idx_kitchen_id (kitchen_id) のみ
-- クエリ: SELECT * FROM kc_cooking_tasks WHERE kitchen_id = ? ORDER BY started_at ASC
-- 問題: kitchen_id で絞り込んだ後に started_at で filesort が発生する
--
-- 改善: (kitchen_id, started_at) の複合インデックスにすることで
-- kitchen_id の等値条件 + started_at のソートをインデックスだけで解決。
-- filesort が不要になる。
-- 元の idx_kitchen_id は新しい複合インデックスの左端プレフィックスでカバーされるため削除。
DROP INDEX idx_kitchen_id ON kc_cooking_tasks;
CREATE INDEX idx_kitchen_started ON kc_cooking_tasks (kitchen_id, started_at ASC);

-- クエリ: UPDATE kc_cooking_tasks SET status = ? WHERE kitchen_id = ? AND order_id = ?
-- 現状: kitchen_id と order_id の複合条件に対応するインデックスがない。
-- kitchen_id の等値 + order_id の等値なので、複合インデックスで両方カバー。
-- idx_order_id (order_id) 単独は残す（order_id だけでの検索もあるため）。
CREATE INDEX idx_kitchen_order ON kc_cooking_tasks (kitchen_id, order_id);
