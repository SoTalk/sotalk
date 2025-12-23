-- Rollback: Add display_name column back to users table
-- This migration restores the display_name column if needed

BEGIN;

-- Add the display_name column back
ALTER TABLE users ADD COLUMN IF NOT EXISTS display_name VARCHAR(100);

-- Create unique index on display_name
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_display_name ON users(display_name);

-- Copy username to display_name for existing users
UPDATE users SET display_name = username WHERE display_name IS NULL;

-- Make display_name NOT NULL after copying data
ALTER TABLE users ALTER COLUMN display_name SET NOT NULL;

COMMIT;
