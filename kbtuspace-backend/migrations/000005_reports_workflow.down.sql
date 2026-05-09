DROP INDEX IF EXISTS idx_reports_reporter_id;
DROP INDEX IF EXISTS idx_reports_target_post_id;
DROP INDEX IF EXISTS idx_reports_status_created_at;

ALTER TABLE reports
    DROP CONSTRAINT IF EXISTS reports_target_type_check;

ALTER TABLE reports
    DROP CONSTRAINT IF EXISTS reports_status_check;

ALTER TABLE reports
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS review_note,
    DROP COLUMN IF EXISTS target_type;
