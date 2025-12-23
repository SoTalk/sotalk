package postgres

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	LogLevel        logger.LogLevel
}

// NewDatabase creates a new database connection
func NewDatabase(config DatabaseConfig) (*gorm.DB, error) {
	// Set default log level
	if config.LogLevel == 0 {
		config.LogLevel = logger.Info
	}

	// GORM configuration
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true, // Prepared statement cache
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(config.DSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database connected successfully")

	return db, nil
}

// AutoMigrate runs auto migration for all models
func AutoMigrate(db *gorm.DB) error {
	log.Println("🔄 Running database auto-migration...")

	// Drop display_name column if it exists (migration to username-only)
	if err := dropDisplayNameColumn(db); err != nil {
		log.Printf("⚠️  Display name migration warning: %v", err)
		// Don't fail if migration already ran
	}

	// First, run manual SQL migration for referral system
	// This ensures existing users get referral codes before GORM creates indexes
	if err := runReferralMigration(db); err != nil {
		log.Printf("⚠️  Referral migration warning: %v", err)
		// Don't fail if migration already ran
	}

	// Add all models here for auto-migration
	err := db.AutoMigrate(
		&User{},
		&Conversation{},
		&ConversationParticipant{},
		&Message{},
		// Group models (Day 5)
		&Group{},
		&GroupMember{},
		// Channel models (Day 6)
		&Channel{},
		&ChannelSubscriber{},
		&ChannelAdmin{},
		// Media models (Day 7)
		&Media{},
		// Wallet models (Day 9)
		&Wallet{},
		&Transaction{},
		// Payment models (Day 10)
		&PaymentRequest{},
		// Privacy & Security models (Day 12)
		&PrivacySettings{},
		&BlockedUser{},
		&DisappearingMessagesConfig{},
		&TwoFactorAuth{},
		// Advanced Messaging models (Day 13)
		&MessageReaction{},
		&PinnedMessage{},
		&ForwardedMessage{},
		&MessageMention{},
		&Status{},
		&StatusView{},
		&Contact{},
		&ContactInvite{},
		// Notification models
		&Notification{},
		&NotificationSettings{},
		// Referral models
		&Referral{},
	)

	if err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	log.Println("✅ Auto-migration completed successfully")
	return nil
}

// dropDisplayNameColumn drops the display_name column from users table
func dropDisplayNameColumn(db *gorm.DB) error {
	// Check if display_name column exists
	var columnExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'users' AND column_name = 'display_name'
		)
	`).Scan(&columnExists).Error

	if err != nil {
		return err
	}

	// If column exists, migrate data and drop it
	if columnExists {
		log.Println("📝 Migrating from display_name to username...")

		// Update users with null or empty username to use display_name
		if err := db.Exec(`
			UPDATE users
			SET username = COALESCE(NULLIF(username, ''), display_name, SUBSTRING(wallet_address FROM 1 FOR 8))
			WHERE username IS NULL OR username = ''
		`).Error; err != nil {
			log.Printf("⚠️  Warning updating usernames: %v", err)
		}

		// Drop the unique index on display_name first
		if err := db.Exec("DROP INDEX IF EXISTS idx_users_display_name").Error; err != nil {
			log.Printf("⚠️  Warning dropping index: %v", err)
		}

		// Drop the display_name column
		if err := db.Exec("ALTER TABLE users DROP COLUMN IF EXISTS display_name").Error; err != nil {
			return err
		}

		log.Println("✅ Dropped display_name column successfully")
	} else {
		log.Println("ℹ️  display_name column already removed")
	}

	// Ensure all existing users have a username (in case they don't)
	log.Println("📝 Ensuring all users have usernames...")
	if err := db.Exec(`
		UPDATE users
		SET username = SUBSTRING(wallet_address FROM 1 FOR 8)
		WHERE username IS NULL OR username = ''
	`).Error; err != nil {
		log.Printf("⚠️  Warning ensuring usernames: %v", err)
	}

	return nil
}

// runReferralMigration runs the referral system migration
func runReferralMigration(db *gorm.DB) error {
	// Check if referral_code column exists
	var columnExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'users' AND column_name = 'referral_code'
		)
	`).Scan(&columnExists).Error

	if err != nil {
		return err
	}

	// If column doesn't exist, add it
	if !columnExists {
		log.Println("📝 Adding referral_code column to users table...")
		if err := db.Exec("ALTER TABLE users ADD COLUMN referral_code VARCHAR(8)").Error; err != nil {
			return err
		}
	}

	// Generate referral codes for existing users without codes
	log.Println("📝 Generating referral codes for existing users...")
	err = db.Exec(`
		DO $$
		DECLARE
			user_record RECORD;
			new_code VARCHAR(8);
			code_exists BOOLEAN;
		BEGIN
			FOR user_record IN SELECT id FROM users WHERE referral_code IS NULL OR referral_code = ''
			LOOP
				LOOP
					new_code := UPPER(SUBSTRING(ENCODE(gen_random_bytes(5), 'base32') FROM 1 FOR 8));
					SELECT EXISTS(SELECT 1 FROM users WHERE referral_code = new_code) INTO code_exists;
					EXIT WHEN NOT code_exists;
				END LOOP;
				UPDATE users SET referral_code = new_code WHERE id = user_record.id;
			END LOOP;
		END $$;
	`).Error

	if err != nil {
		return err
	}

	// Check if referrals table exists
	var tableExists bool
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_name = 'referrals'
		)
	`).Scan(&tableExists).Error

	if err != nil {
		return err
	}

	// Create referrals table if it doesn't exist
	if !tableExists {
		log.Println("📝 Creating referrals table...")
		// Let GORM AutoMigrate create the referrals table
		// It will be created in the main AutoMigrate call
	}

	log.Println("✅ Referral migration completed")
	return nil
}

