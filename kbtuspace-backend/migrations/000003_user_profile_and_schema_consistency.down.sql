DROP INDEX IF EXISTS idx_registrations_event_status;
DROP INDEX IF EXISTS idx_events_feed_filter;
DROP INDEX IF EXISTS idx_posts_feed_filter;

ALTER TABLE posts
    DROP COLUMN IF EXISTS current_count,
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE registrations
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE users
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS is_banned;
