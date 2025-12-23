-- Update pinned_messages table to allow multiple users to pin the same message
-- Change from unique index on message_id to composite unique index on (message_id, pinned_by)

-- Drop the old unique index if it exists
DROP INDEX IF EXISTS idx_pinned_messages_message_id;

-- Create composite unique index to allow multiple users to pin the same message
CREATE UNIQUE INDEX IF NOT EXISTS idx_pinned_message_user ON pinned_messages(message_id, pinned_by);

-- Add comment to explain the change
COMMENT ON INDEX idx_pinned_message_user IS 'Composite unique index allowing multiple users to pin the same message independently';
