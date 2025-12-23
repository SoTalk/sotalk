-- Remove Encryption Tables Migration
-- This migration removes all encryption-related tables and simplifies the messaging system to use plaintext

-- Drop message keys cache table (Signal Protocol)
DROP TABLE IF EXISTS message_keys_cache;

-- Drop signed pre-keys table (Signal Protocol)
DROP TABLE IF EXISTS signed_pre_keys;

-- Drop encryption sessions table completely
DROP TABLE IF EXISTS encryption_sessions;

-- Drop pre-keys table
DROP TABLE IF EXISTS pre_keys;

-- Drop identity keys table
DROP TABLE IF EXISTS identity_keys;

-- Remove encryption_key column from media table
ALTER TABLE media DROP COLUMN IF EXISTS encryption_key;

-- Messages table already stores content as TEXT, no changes needed
-- Messages are now stored as plaintext directly
