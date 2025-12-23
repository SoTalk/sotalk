-- Add referral system to database
-- Migration to add referral_code to users and create referrals table

BEGIN;

-- Drop existing index if it exists (in case of failed migration)
DROP INDEX IF EXISTS idx_users_referral_code;

-- Add referral_code column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS referral_code VARCHAR(8);

-- Update existing users with random referral codes
-- Generate unique 8-character codes for existing users
DO $$
DECLARE
    user_record RECORD;
    new_code VARCHAR(8);
    code_exists BOOLEAN;
BEGIN
    FOR user_record IN SELECT id FROM users WHERE referral_code IS NULL OR referral_code = ''
    LOOP
        -- Generate unique code
        LOOP
            new_code := UPPER(SUBSTRING(ENCODE(gen_random_bytes(5), 'base32') FROM 1 FOR 8));
            -- Check if code already exists
            SELECT EXISTS(SELECT 1 FROM users WHERE referral_code = new_code) INTO code_exists;
            EXIT WHEN NOT code_exists;
        END LOOP;

        -- Update user with unique code
        UPDATE users SET referral_code = new_code WHERE id = user_record.id;
    END LOOP;
END $$;

-- Create unique index on referral_code (after all users have codes)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_referral_code ON users(referral_code) WHERE referral_code IS NOT NULL AND referral_code != '';

-- Create referrals table
CREATE TABLE IF NOT EXISTS referrals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    referrer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    referee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(8) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reward_type VARCHAR(20),
    reward_amount BIGINT,
    reward_tx_sig VARCHAR(88),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(referee_id)
);

-- Create indexes for referrals table
CREATE INDEX IF NOT EXISTS idx_referrals_referrer ON referrals(referrer_id);
CREATE INDEX IF NOT EXISTS idx_referrals_referee ON referrals(referee_id);
CREATE INDEX IF NOT EXISTS idx_referrals_code ON referrals(code);
CREATE INDEX IF NOT EXISTS idx_referrals_status ON referrals(status);
CREATE INDEX IF NOT EXISTS idx_referrals_deleted_at ON referrals(deleted_at);

COMMIT;
