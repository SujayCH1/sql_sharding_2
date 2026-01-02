BEGIN;

-- 1. Convert existing 'draining' values to 'inactive'
UPDATE shards
SET status = 'inactive'
WHERE status = 'draining';

-- 2. Drop old CHECK constraint
ALTER TABLE shards
DROP CONSTRAINT IF EXISTS chk_shard_status;

-- 3. Add new CHECK constraint (active | inactive)
ALTER TABLE shards
ADD CONSTRAINT chk_shard_status
CHECK (status IN ('active', 'inactive'));

-- 4. Set default to 'inactive'
ALTER TABLE shards
ALTER COLUMN status SET DEFAULT 'inactive';

COMMIT;
