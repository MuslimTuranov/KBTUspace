ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE posts
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE registrations
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE posts
    ADD COLUMN IF NOT EXISTS current_count INT NOT NULL DEFAULT 0;

UPDATE users
SET updated_at = created_at
WHERE updated_at IS NULL;

UPDATE posts
SET updated_at = created_at
WHERE updated_at IS NULL;

UPDATE registrations
SET updated_at = created_at
WHERE updated_at IS NULL;

UPDATE posts p
SET current_count = COALESCE((
    SELECT COUNT(*)
    FROM registrations r
    WHERE r.event_id = p.id AND r.status = 'registered'
), 0)
WHERE p.event_date IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_posts_feed_filter
    ON posts (status, scope, faculty_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_events_feed_filter
    ON posts (status, scope, faculty_id, event_date)
    WHERE event_date IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_registrations_event_status
    ON registrations (event_id, status);
