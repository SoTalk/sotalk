-- Rollback Signal Protocol Schema Migration
-- This migration removes all Signal Protocol-related tables and columns

-- Drop message keys cache table
DROP TABLE IF EXISTS message_keys_cache;

-- Drop signed pre-keys table
DROP TABLE IF EXISTS signed_pre_keys;

-- Remove Double Ratchet state columns from encryption_sessions
ALTER TABLE encryption_sessions
DROP COLUMN IF EXISTS root_key,
DROP COLUMN IF EXISTS sending_chain_key,
DROP COLUMN IF EXISTS receiving_chain_key,
DROP COLUMN IF EXISTS sending_chain_length,
DROP COLUMN IF EXISTS receiving_chain_length,
DROP COLUMN IF NOT EXISTS previous_sending_chain_length,
DROP COLUMN IF EXISTS dh_ratchet_public_key,
DROP COLUMN IF EXISTS dh_ratchet_private_key;
