-- Create passkey_credentials table for WebAuthn
CREATE TABLE IF NOT EXISTS passkey_credentials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    credential_id TEXT NOT NULL UNIQUE,
    public_key BYTEA NOT NULL,
    attestation_type TEXT DEFAULT 'none',
    aaguid BYTEA,
    sign_count BIGINT DEFAULT 0,
    transports TEXT[],
    backup_eligible BOOLEAN DEFAULT FALSE,
    backup_state BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    last_used_at TIMESTAMP,

    CONSTRAINT fk_passkey_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_passkey_user_id ON passkey_credentials(user_id);
CREATE INDEX idx_passkey_credential_id ON passkey_credentials(credential_id);
CREATE INDEX idx_passkey_created_at ON passkey_credentials(created_at DESC);

-- Comments
COMMENT ON TABLE passkey_credentials IS 'Stores WebAuthn passkey credentials for users';
COMMENT ON COLUMN passkey_credentials.credential_id IS 'Base64 encoded credential ID from WebAuthn';
COMMENT ON COLUMN passkey_credentials.public_key IS 'Public key in COSE format';
COMMENT ON COLUMN passkey_credentials.sign_count IS 'Sign counter for replay attack prevention';
COMMENT ON COLUMN passkey_credentials.transports IS 'Supported authenticator transports (usb, nfc, ble, internal)';
COMMENT ON COLUMN passkey_credentials.backup_eligible IS 'Whether credential can be backed up';
COMMENT ON COLUMN passkey_credentials.backup_state IS 'Whether credential is currently backed up';
