ALTER TABLE projects
ADD COLUMN IF NOT EXISTS status BOOLEAN NOT NULL DEFAULT false;

UPDATE projects SET status = false WHERE status IS NULL;
