-- Rollback migration (recreate encryption tables if needed)
-- Note: This will recreate empty tables, but won't restore any data

-- Add back encryption_key column to media table
ALTER TABLE media ADD COLUMN IF NOT EXISTS encryption_key BYTEA;

-- Recreate identity_keys table
CREATE TABLE IF NOT EXISTS identity_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    public_key BYTEA NOT NULL,
    private_key BYTEA NOT NULL,
    signing_key BYTEA NOT NULL,
    verifying_key BYTEA NOT NULL,
    key_fingerprint VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_identity_keys_user ON identity_keys(user_id);

-- Recreate pre_keys table
CREATE TABLE IF NOT EXISTS pre_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_id INT NOT NULL,
    public_key BYTEA NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, key_id)
);

CREATE INDEX idx_pre_keys_user ON pre_keys(user_id);

-- Recreate encryption_sessions table
CREATE TABLE IF NOT EXISTS encryption_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    peer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_key BYTEA NOT NULL,
    ratchet_key BYTEA,
    message_counter INT NOT NULL DEFAULT 0,
    root_key BYTEA,
    sending_chain_key BYTEA,
    receiving_chain_key BYTEA,
    sending_chain_length INT DEFAULT 0,
    receiving_chain_length INT DEFAULT 0,
    previous_sending_chain_length INT DEFAULT 0,
    dh_ratchet_public_key BYTEA,
    dh_ratchet_private_key BYTEA,
    last_message_at TIMESTAMP,
    established_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, peer_id)
);

CREATE INDEX idx_encryption_sessions_user ON encryption_sessions(user_id);
CREATE INDEX idx_encryption_sessions_peer ON encryption_sessions(peer_id);

-- Recreate signed_pre_keys table
CREATE TABLE IF NOT EXISTS signed_pre_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_id INT NOT NULL,
    public_key BYTEA NOT NULL,
    private_key BYTEA NOT NULL,
    signature BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, key_id)
);

CREATE INDEX idx_signed_pre_keys_user ON signed_pre_keys(user_id);

-- Recreate message_keys_cache table
CREATE TABLE IF NOT EXISTS message_keys_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES encryption_sessions(id) ON DELETE CASCADE,
    chain_key BYTEA NOT NULL,
    message_number INT NOT NULL,
    message_key BYTEA NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, chain_key, message_number)
);

CREATE INDEX idx_message_keys_cache_session ON message_keys_cache(session_id);
CREATE INDEX idx_message_keys_cache_expires ON message_keys_cache(expires_at);
