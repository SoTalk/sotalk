-- Rollback: Revert to single unique index on message_id
-- This will prevent multiple users from pinning the same message

-- Drop the composite unique index
DROP INDEX IF EXISTS idx_pinned_message_user;

-- Recreate the old unique index on message_id alone
CREATE UNIQUE INDEX IF NOT EXISTS idx_pinned_messages_message_id ON pinned_messages(message_id);
