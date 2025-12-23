-- Signal Protocol Schema Migration
-- This migration adds support for the full Signal Protocol implementation
-- including Double Ratchet Algorithm and X3DH key agreement

-- Signed pre-keys (different from one-time pre-keys)
-- These are long-lived keys signed by the identity key
CREATE TABLE IF NOT EXISTS signed_pre_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_id INT NOT NULL,
    public_key BYTEA NOT NULL,
    private_key BYTEA NOT NULL, -- Encrypted at rest with master key
    signature BYTEA NOT NULL, -- Signed by identity key (Ed25519 signature)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, key_id)
);

-- Index for faster lookups
CREATE INDEX idx_signed_pre_keys_user ON signed_pre_keys(user_id);

-- Update encryption_sessions table to include Double Ratchet state
-- These columns store the state of the Double Ratchet Algorithm
ALTER TABLE encryption_sessions
ADD COLUMN IF NOT EXISTS root_key BYTEA, -- Root chain key for DH ratchet
ADD COLUMN IF NOT EXISTS sending_chain_key BYTEA, -- Current sending chain key
ADD COLUMN IF NOT EXISTS receiving_chain_key BYTEA, -- Current receiving chain key
ADD COLUMN IF NOT EXISTS sending_chain_length INT DEFAULT 0, -- Number of messages sent in current chain
ADD COLUMN IF NOT EXISTS receiving_chain_length INT DEFAULT 0, -- Number of messages received in current chain
ADD COLUMN IF NOT EXISTS previous_sending_chain_length INT DEFAULT 0, -- For out-of-order message handling
ADD COLUMN IF NOT EXISTS dh_ratchet_public_key BYTEA, -- Current DH ratchet public key (X25519)
ADD COLUMN IF NOT EXISTS dh_ratchet_private_key BYTEA; -- Current DH ratchet private key (X25519, encrypted at rest)

-- Message keys cache for out-of-order message handling
-- When messages arrive out of order, we cache the derived message keys
-- so they can be decrypted when they arrive
CREATE TABLE IF NOT EXISTS message_keys_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES encryption_sessions(id) ON DELETE CASCADE,
    chain_key BYTEA NOT NULL, -- The chain key this message key was derived from
    message_number INT NOT NULL, -- Message number in the chain
    message_key BYTEA NOT NULL, -- The derived message key (ChaCha20-Poly1305 key)
    expires_at TIMESTAMP NOT NULL, -- Auto-expire after 7 days for security
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, chain_key, message_number)
);

-- Indexes for efficient queries
CREATE INDEX idx_message_keys_cache_session ON message_keys_cache(session_id);
CREATE INDEX idx_message_keys_cache_expires ON message_keys_cache(expires_at);

-- Comments for documentation
COMMENT ON TABLE signed_pre_keys IS 'Signed pre-keys for X3DH key agreement protocol';
COMMENT ON TABLE message_keys_cache IS 'Cache for message keys to handle out-of-order messages in Double Ratchet';
COMMENT ON COLUMN encryption_sessions.root_key IS 'Root key for DH ratchet (KDF_RK)';
COMMENT ON COLUMN encryption_sessions.sending_chain_key IS 'Current sending chain key (KDF_CK)';
COMMENT ON COLUMN encryption_sessions.receiving_chain_key IS 'Current receiving chain key (KDF_CK)';
COMMENT ON COLUMN encryption_sessions.dh_ratchet_public_key IS 'Current DH ratchet public key (Diffie-Hellman key)';
