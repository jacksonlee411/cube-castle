-- Plan 402D · 恢复 legacy 表写锁
INSERT INTO plan402_runtime_flags (id, enforce_legacy_lock)
VALUES (1, true)
ON CONFLICT (id) DO NOTHING;

UPDATE plan402_runtime_flags
SET enforce_legacy_lock = true,
    updated_at = now()
WHERE id = 1;

SELECT enforce_legacy_lock AS legacy_lock_enabled,
       updated_at
FROM plan402_runtime_flags
WHERE id = 1;
