ALTER TABLE posts
    DROP COLUMN IF EXISTS rejection_reason,
    DROP COLUMN IF EXISTS approved_at,
    DROP COLUMN IF EXISTS approved_by,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS scope;
