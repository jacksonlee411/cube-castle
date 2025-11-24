-- Plan 402D · 双时态回放辅助脚本
-- 作用：在给定 transaction_timestamp 的情况下导出 standard_object_versions 快照，
--        用于与 `cmd/tools/standardobject-validator` 报告对账。
--
-- 用法：
--   psql "$DATABASE_URL" \
--     -v as_of_ts="'2025-11-30 10:00:00+00'" \
--     -f database/scripts/plan402/checkpoints/replay_view.sql \
--     | tee logs/plan402/verification/as-of-20251130T100000Z.log
--
-- 若未提供 as_of_ts，将默认使用 now()。

\if :{?as_of_ts}
\else
\set as_of_ts '''now()'''
\endif

\echo '🕒 Plan 402D · 回放 transaction timestamp: ' :as_of_ts

DROP TABLE IF EXISTS tmp_plan402_replay CASCADE;

CREATE TEMP TABLE tmp_plan402_replay AS
SELECT
  so.object_type,
  so.code,
  so.display_name,
  so.tenant_code,
  so.status,
  so.labels,
  sov.version_code,
  sov.effective_date,
  sov.end_date,
  sov.is_current,
  sov.payload,
  sov.audit,
  lower(sov.transaction_range)  AS transaction_from,
  upper(sov.transaction_range)  AS transaction_to
FROM standard_object_versions sov
JOIN standard_objects so ON so.id = sov.object_id
WHERE lower(sov.transaction_range) <= (:as_of_ts)::timestamptz
  AND (
    upper(sov.transaction_range) IS NULL
    OR upper(sov.transaction_range) > (:as_of_ts)::timestamptz
  )
ORDER BY so.object_type, so.code, sov.effective_date;

\echo '📄 回放结果（每行代表一个 SOM 版本）：'
TABLE tmp_plan402_replay;

\echo '📊 统计摘要：'
SELECT
  object_type,
  COUNT(*)              AS version_count,
  COUNT(*) FILTER (WHERE is_current) AS current_versions,
  MIN(effective_date)   AS earliest_effective_date,
  MAX(effective_date)   AS latest_effective_date
FROM tmp_plan402_replay
GROUP BY object_type
ORDER BY object_type;

DROP TABLE IF EXISTS tmp_plan402_replay CASCADE;
