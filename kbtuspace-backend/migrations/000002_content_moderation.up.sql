ALTER TABLE posts
    ADD COLUMN IF NOT EXISTS scope VARCHAR(20) NOT NULL DEFAULT 'faculty',
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'approved',
    ADD COLUMN IF NOT EXISTS approved_by INT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS rejection_reason TEXT;

UPDATE posts
SET scope = CASE
        WHEN faculty_id IS NULL THEN 'global'
        ELSE 'faculty'
    END,
    status = 'approved'
WHERE scope IS NULL
   OR status IS NULL;
