-- Drop display_name column from users table
-- Migration to consolidate user identity to use only username

BEGIN;

-- Drop the unique index on display_name first
DROP INDEX IF EXISTS idx_users_display_name;

-- Drop the display_name column
ALTER TABLE users DROP COLUMN IF EXISTS display_name;

COMMIT;
