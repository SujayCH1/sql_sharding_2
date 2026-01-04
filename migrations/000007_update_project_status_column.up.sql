-- 1. Normalize existing data
UPDATE projects
SET status = LOWER(status);

-- 2. Ensure only valid values exist
UPDATE projects
SET status = 'inactive'
WHERE status NOT IN ('active', 'inactive');

-- 3. Drop existing default
ALTER TABLE projects
ALTER COLUMN status DROP DEFAULT;

-- 4. Add CHECK constraint
ALTER TABLE projects
ADD CONSTRAINT chk_project_status
CHECK (status IN ('active', 'inactive'));

-- 5. Set default to inactive
ALTER TABLE projects
ALTER COLUMN status SET DEFAULT 'inactive';
