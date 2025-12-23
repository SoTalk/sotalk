package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpDelivery "github.com/yourusername/sotalk/internal/delivery/http"
	"github.com/yourusername/sotalk/internal/delivery/http/handler"
	"github.com/yourusername/sotalk/internal/delivery/websocket"
	"github.com/yourusername/sotalk/internal/infrastructure/solana"
	"github.com/yourusername/sotalk/internal/infrastructure/storage"
	"github.com/yourusername/sotalk/internal/repository/postgres"
	redisRepo "github.com/yourusername/sotalk/internal/repository/redis"
	"github.com/yourusername/sotalk/internal/usecase/auth"
	"github.com/yourusername/sotalk/internal/usecase/channel"
	"github.com/yourusername/sotalk/internal/usecase/contact"
	"github.com/yourusername/sotalk/internal/usecase/group"
	"github.com/yourusername/sotalk/internal/usecase/media"
	"github.com/yourusername/sotalk/internal/usecase/message"
	"github.com/yourusername/sotalk/internal/usecase/notification"
	"github.com/yourusername/sotalk/internal/usecase/payment"
	"github.com/yourusername/sotalk/internal/usecase/privacy"
	"github.com/yourusername/sotalk/internal/usecase/referral"
	"github.com/yourusername/sotalk/internal/usecase/status"
	"github.com/yourusername/sotalk/internal/usecase/user"
	walletUseCase "github.com/yourusername/sotalk/internal/usecase/wallet"
	"github.com/yourusername/sotalk/pkg/config"
	"github.com/yourusername/sotalk/pkg/email"
	"github.com/yourusername/sotalk/pkg/logger"
	"github.com/yourusername/sotalk/pkg/middleware"
	pkgRedis "github.com/yourusername/sotalk/pkg/redis"
	"go.uber.org/zap"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Server.Environment); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Soldef API Server",
		zap.String("environment", cfg.Server.Environment),
		zap.Int("port", cfg.Server.Port),
	)

	// Initialize database
	dbConfig := postgres.DatabaseConfig{
		DSN:             cfg.Database.GetDSN(),
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		LogLevel:        gormLogger.Info,
	}

	db, err := postgres.NewDatabase(dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run auto-migration on startup
	logger.Info("Running database migrations...")
	if err := postgres.AutoMigrate(db); err != nil {
		logger.Fatal("Failed to run auto-migration", zap.Error(err))
	}
	logger.Info("✅ Database migrations completed")

	// Initialize Redis client for repository caches
	redisClient, err := redisRepo.NewClient(redisRepo.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Info("✅ Redis connected successfully")

	// Initialize pkg/redis client for auth challenge caching
	authRedisClient, err := pkgRedis.NewClient(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis for auth", zap.Error(err))
	}
	logger.Info("✅ Auth Redis client initialized")

	// Initialize Redis caches
	sessionCache := redisRepo.NewSessionCache(redisClient.GetClient())
	messageCache := redisRepo.NewMessageCache(redisClient.GetClient())
	presenceCache := redisRepo.NewPresenceCache(redisClient.GetClient())
	conversationCache := redisRepo.NewConversationCache(redisClient.GetClient())
	logger.Info("✅ Redis caches initialized",
		zap.String("session_cache", "ready"),
		zap.String("message_cache", "ready"),
		zap.String("presence_cache", "ready"),
		zap.String("conversation_cache", "ready"),
	)

	// Caches are available for service integration
	// - sessionCache: Use in auth service for session management
	// - messageCache: Use in message service for fast message retrieval
	// - presenceCache: Available for presence tracking integration
	// - conversationCache: Use in conversation service for conversation lists
	// See internal/repository/redis/README.md for integration examples
	_ = sessionCache      // Available for auth service integration
	_ = messageCache      // Available for message service integration
	_ = presenceCache     // Available for presence tracking integration
	_ = conversationCache // Available for conversation service integration

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	messageRepo := postgres.NewMessageRepository(db)
	// Note: conversationRepo is created with nil presenceChecker initially
	// After wsHub is created, it will be used for online status checking
	conversationRepo := postgres.NewConversationRepository(db, nil)
	groupRepo := postgres.NewGroupRepository(db)
	channelRepo := postgres.NewChannelRepository(db)
	mediaRepo := postgres.NewMediaRepository(db)
	walletRepo := postgres.NewWalletRepository(db)
	paymentRepo := postgres.NewPaymentRequestRepository(db)
	privacyRepo := postgres.NewPrivacyRepository(db)
	statusRepo := postgres.NewStatusRepository(db)         // Day 13
	contactRepo := postgres.NewContactRepository(db)       // Day 13
	notificationRepo := postgres.NewNotificationRepository(db) // Notifications
	referralRepo := postgres.NewReferralRepository(db)     // Referral system

	// Initialize JWT manager
	jwtManager := middleware.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// Initialize crypto service (Day 11)
	// Use JWT secret as master key for encrypting keys at rest (derive 32 bytes)
	masterKey := []byte(cfg.JWT.Secret)
	if len(masterKey) < 32 {
		// Pad to 32 bytes if needed
		padded := make([]byte, 32)
		copy(padded, masterKey)
		masterKey = padded
	} else if len(masterKey) > 32 {
		masterKey = masterKey[:32]
	}
	logger.Info("✅ Crypto service initialized")

	// Initialize Solana RPC client (Day 9)
	// NOTE: Moved before auth service so it can fetch wallet balances during registration
	solanaClient, err := solana.NewClient(solana.Config{
		RPCEndpoint: cfg.Solana.RPCEndpoint,
		Network:     cfg.Solana.Network,
	})
	if err != nil {
		logger.Fatal("Failed to initialize Solana client", zap.Error(err))
	}
	logger.Info("✅ Solana RPC client initialized")

	// Initialize referral service (needed by auth service)
	referralService := referral.NewService(
		referralRepo,
		userRepo,
	)
	logger.Info("✅ Referral service initialized")

	// Initialize services (use cases)
	authService := auth.NewService(
		userRepo,
		walletRepo,
		jwtManager,
		solanaClient,
		authRedisClient,
		referralService,
	)

	channelService := channel.NewService(
		channelRepo,
		conversationRepo,
		userRepo,
	)

	// Initialize storage service (Azure Blob or Local)
	var storageService storage.Service

	if cfg.Storage.Provider == "azure" {
		storageService, err = storage.NewAzureBlobService(storage.AzureBlobConfig{
			AccountName:   cfg.Storage.AzureAccountName,
			AccountKey:    cfg.Storage.AzureAccountKey,
			ContainerName: cfg.Storage.AzureContainer,
		})
		if err != nil {
			logger.Fatal("Failed to initialize Azure Blob Storage", zap.Error(err))
		}
		logger.Info("✅ Azure Blob Storage initialized")
	} else {
		storageService, err = storage.NewLocalStorage(storage.LocalStorageConfig{
			BasePath: cfg.Storage.LocalPath,
			BaseURL:  "/uploads",
		})
		if err != nil {
			logger.Fatal("Failed to initialize Local Storage", zap.Error(err))
		}
		logger.Info("✅ Local Storage initialized")
	}

	mediaService := media.NewService(
		mediaRepo,
		userRepo,
		storageService,
	)

	// Initialize email client
	emailClient := email.NewClient(email.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		FromName: cfg.SMTP.FromName,
	})
	logger.Info("✅ Email client initialized",
		zap.String("host", cfg.SMTP.Host),
		zap.Int("port", cfg.SMTP.Port),
	)

	walletService := walletUseCase.NewService(
		walletRepo,
		userRepo,
		solanaClient,
	)

	// Initialize WebSocket hub and start it
	wsHub := websocket.NewHub(conversationRepo, userRepo)
	go wsHub.Run()
	logger.Info("✅ WebSocket Hub started")

	// Recreate conversationRepo with Hub as presence checker
	// This enables online status checking when fetching participants
	conversationRepo = postgres.NewConversationRepository(db, wsHub)
	logger.Info("✅ Conversation repository updated with presence checking")

	// Create WebSocket broadcaster
	wsBroadcaster := websocket.NewBroadcaster(wsHub)
	logger.Info("✅ WebSocket broadcaster initialized")

	// Initialize group service with WebSocket support
	groupService := group.NewService(
		groupRepo,
		conversationRepo,
		userRepo,
		wsBroadcaster,
	)
	logger.Info("✅ Group service initialized with WebSocket support")

	// Initialize message service with WebSocket support
	messageService := message.NewService(
		messageRepo,
		conversationRepo,
		userRepo,
		wsBroadcaster,
	)
	logger.Info("✅ Message service initialized with WebSocket support")

	paymentService := payment.NewService(
		paymentRepo,
		userRepo,
		walletRepo,
		solanaClient,
		wsBroadcaster,
	)
	logger.Info("✅ Payment service initialized with WebSocket support")


	// Initialize privacy service (Day 12)
	privacyService := privacy.NewService(privacyRepo, conversationRepo)
	logger.Info("✅ Privacy service initialized")

	// Initialize status service (Day 13)
	statusService := status.NewService(statusRepo, contactRepo, userRepo)
	logger.Info("✅ Status service initialized")

	// Initialize contact service (Day 13)
	contactService := contact.NewService(contactRepo)
	logger.Info("✅ Contact service initialized")

	// Initialize user service
	userService := user.NewService(userRepo, emailClient)
	logger.Info("✅ User service initialized")

	// Initialize notification service
	notificationService := notification.NewService(notificationRepo)
	logger.Info("✅ Notification service initialized")


	// Initialize HTTP handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	messageHandler := handler.NewMessageHandler(messageService, privacyRepo, contactRepo)
	groupHandler := handler.NewGroupHandler(groupService)
	channelHandler := handler.NewChannelHandler(channelService)
	mediaHandler := handler.NewMediaHandler(mediaService)
	walletHandler := handler.NewWalletHandler(walletService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	privacyHandler := handler.NewPrivacyHandler(privacyService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	statusHandler := handler.NewStatusHandler(statusService)    // Day 13
	contactHandler := handler.NewContactHandler(contactService) // Day 13
	referralHandler := handler.NewReferralHandler(referralService)
	wsHandler := websocket.NewHandler(wsHub, messageService)    // WebSocket handler
	logger.Info("✅ WebSocket handler initialized")

	// Initialize HTTP router
	router := httpDelivery.NewRouter(authHandler, userHandler, messageHandler, groupHandler, channelHandler, mediaHandler, walletHandler, paymentHandler, privacyHandler, notificationHandler, statusHandler, contactHandler, referralHandler, wsHandler, jwtManager)
	ginEngine := router.Setup(cfg.Server.Environment)
	logger.Info("✅ HTTP router configured with WebSocket routes")

	// Create HTTP server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      ginEngine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting",
			zap.String("address", serverAddr),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	logger.Info("✅ Server started successfully",
		zap.String("address", serverAddr),
		zap.String("health_check", fmt.Sprintf("http://localhost:%d/health", cfg.Server.Port)),
	)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	// Close database connection
	sqlDB, _ := db.DB()
	if err := sqlDB.Close(); err != nil {
		logger.Error("Failed to close database connection", zap.Error(err))
	}

	// Close Redis connections
	if err := redisClient.Close(); err != nil {
		logger.Error("Failed to close Redis connection", zap.Error(err))
	}
	if err := authRedisClient.Close(); err != nil {
		logger.Error("Failed to close auth Redis connection", zap.Error(err))
	}

	logger.Info("✅ Server exited successfully")
}
