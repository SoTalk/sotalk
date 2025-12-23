-- Rollback: Remove referral system from database

BEGIN;

-- Drop referrals table
DROP TABLE IF EXISTS referrals;

-- Drop referral_code column from users table
ALTER TABLE users DROP COLUMN IF EXISTS referral_code;

-- Drop index
DROP INDEX IF EXISTS idx_users_referral_code;

COMMIT;
