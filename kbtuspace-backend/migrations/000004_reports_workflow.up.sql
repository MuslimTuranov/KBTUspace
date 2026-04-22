ALTER TABLE reports
    ADD COLUMN IF NOT EXISTS target_type VARCHAR(20) NOT NULL DEFAULT 'post',
    ADD COLUMN IF NOT EXISTS review_note TEXT,
    ADD COLUMN IF NOT EXISTS reviewed_by INT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

UPDATE reports
SET target_type = CASE
                      WHEN EXISTS (
                          SELECT 1
                          FROM posts p
                          WHERE p.id = reports.target_post_id
                            AND p.event_date IS NOT NULL
                      ) THEN 'event'
                      ELSE 'post'
    END
WHERE target_type IS NULL OR target_type = '';

UPDATE reports
SET updated_at = created_at
WHERE updated_at IS NULL;

ALTER TABLE reports
DROP CONSTRAINT IF EXISTS reports_status_check;

ALTER TABLE reports
    ADD CONSTRAINT reports_status_check
        CHECK (status IN ('pending', 'closed', 'rejected'));

ALTER TABLE reports
DROP CONSTRAINT IF EXISTS reports_target_type_check;

ALTER TABLE reports
    ADD CONSTRAINT reports_target_type_check
        CHECK (target_type IN ('post', 'event'));

CREATE INDEX IF NOT EXISTS idx_reports_status_created_at
    ON reports(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_reports_target_post_id
    ON reports(target_post_id);

CREATE INDEX IF NOT EXISTS idx_reports_reporter_id
    ON reports(reporter_id);