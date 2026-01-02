BEGIN;

-- 1. Drop new CHECK constraint
ALTER TABLE shards
DROP CONSTRAINT IF EXISTS chk_shard_status;

-- 2. Restore old CHECK constraint
ALTER TABLE shards
ADD CONSTRAINT chk_shard_status
CHECK (status IN ('active', 'draining', 'offline'));

-- 3. Remove default value
ALTER TABLE shards
ALTER COLUMN status DROP DEFAULT;

COMMIT;
