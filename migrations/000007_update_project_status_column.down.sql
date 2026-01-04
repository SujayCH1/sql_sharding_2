-- 1. Drop CHECK constraint
ALTER TABLE projects
DROP CONSTRAINT IF EXISTS chk_project_status;

-- 2. Drop current default
ALTER TABLE projects
ALTER COLUMN status DROP DEFAULT;

-- 3. Restore previous default (ACTIVE)
ALTER TABLE projects
ALTER COLUMN status SET DEFAULT 'ACTIVE';
