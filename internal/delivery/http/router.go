package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/sotalk/internal/delivery/http/handler"
	httpMiddleware "github.com/yourusername/sotalk/internal/delivery/http/middleware"
	"github.com/yourusername/sotalk/internal/delivery/websocket"
	"github.com/yourusername/sotalk/pkg/logger"
	"github.com/yourusername/sotalk/pkg/middleware"
	"go.uber.org/zap"
)

// Router holds all HTTP handlers
type Router struct {
	healthHandler       *handler.HealthHandler
	authHandler         *handler.AuthHandler
	userHandler         *handler.UserHandler
	messageHandler      *handler.MessageHandler
	groupHandler        *handler.GroupHandler
	channelHandler      *handler.ChannelHandler
	mediaHandler        *handler.MediaHandler
	walletHandler       *handler.WalletHandler
	paymentHandler      *handler.PaymentHandler
	privacyHandler      *handler.PrivacyHandler
	notificationHandler *handler.NotificationHandler
	statusHandler       *handler.StatusHandler  // Day 13
	contactHandler      *handler.ContactHandler // Day 13
	referralHandler     *handler.ReferralHandler
	wsHandler           *websocket.Handler // WebSocket handler (Day 4)
	jwtManager          *middleware.JWTManager
}

// NewRouter creates a new router instance
func NewRouter(authHandler *handler.AuthHandler, userHandler *handler.UserHandler, messageHandler *handler.MessageHandler, groupHandler *handler.GroupHandler, channelHandler *handler.ChannelHandler, mediaHandler *handler.MediaHandler, walletHandler *handler.WalletHandler, paymentHandler *handler.PaymentHandler, privacyHandler *handler.PrivacyHandler, notificationHandler *handler.NotificationHandler, statusHandler *handler.StatusHandler, contactHandler *handler.ContactHandler, referralHandler *handler.ReferralHandler, wsHandler *websocket.Handler, jwtManager *middleware.JWTManager) *Router {
	return &Router{
		healthHandler:       handler.NewHealthHandler(),
		authHandler:         authHandler,
		userHandler:         userHandler,
		messageHandler:      messageHandler,
		groupHandler:        groupHandler,
		channelHandler:      channelHandler,
		mediaHandler:        mediaHandler,
		walletHandler:       walletHandler,
		paymentHandler:      paymentHandler,
		privacyHandler:      privacyHandler,
		notificationHandler: notificationHandler,
		statusHandler:       statusHandler,
		contactHandler:      contactHandler,
		referralHandler:     referralHandler,
		wsHandler:           wsHandler,
		jwtManager:          jwtManager,
	}
}

// Setup configures all routes
func (r *Router) Setup(mode string) *gin.Engine {
	// Set Gin mode
	if mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin engine
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware())
	router.Use(CORSMiddleware())

	// Health check routes (no auth required)
	router.GET("/health", r.healthHandler.Health)
	router.GET("/ready", r.healthHandler.Ready)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", r.authHandler.GenerateWallet)
			auth.POST("/challenge", r.authHandler.GenerateChallenge)
			auth.POST("/verify", r.authHandler.VerifySignature)
			auth.POST("/refresh", r.authHandler.RefreshToken)
		}

		// Protected routes (require authentication via header)
		protected := v1.Group("")
		protected.Use(httpMiddleware.AuthMiddleware(r.jwtManager))
		{
			// Auth protected routes
			protected.GET("/auth/me", r.authHandler.GetMe)
			protected.POST("/auth/logout", r.authHandler.Logout)
			protected.DELETE("/auth/account", r.authHandler.DeleteAccount)

			// WebSocket routes (Day 4)
			protected.GET("/ws", r.wsHandler.HandleWebSocket)
			protected.GET("/ws/stats", r.wsHandler.Stats)

			// Message routes (Day 3)
			messages := protected.Group("/messages")
			{
				messages.POST("", r.messageHandler.SendMessage)
				messages.GET("", r.messageHandler.GetMessages)
				messages.POST("/read", r.messageHandler.MarkAsRead)
				messages.PUT("/:id", r.messageHandler.EditMessage)
				messages.DELETE("/:id", r.messageHandler.DeleteMessage)
			}

			// Conversation routes (Day 3, Day 8)
			conversations := protected.Group("/conversations")
			{
				conversations.GET("", r.messageHandler.GetConversations)
				conversations.POST("", r.messageHandler.CreateConversation)
				conversations.GET("/:id/media", r.mediaHandler.GetConversationMedia) // Day 8: Shared media
				conversations.PUT("/:id/archive", r.messageHandler.ArchiveConversation)
				conversations.PUT("/:id/unarchive", r.messageHandler.UnarchiveConversation)
				conversations.DELETE("/:id", r.messageHandler.DeleteConversation)
			}

			// Group routes (Day 5)
			groups := protected.Group("/groups")
			{
				groups.POST("", r.groupHandler.CreateGroup)
				groups.GET("", r.groupHandler.GetUserGroups)
				groups.GET("/:id", r.groupHandler.GetGroup)
				groups.PUT("/:id", r.groupHandler.UpdateGroup)
				groups.PUT("/:id/settings", r.groupHandler.UpdateGroupSettings)
				groups.DELETE("/:id", r.groupHandler.DeleteGroup)
				groups.POST("/:id/members", r.groupHandler.AddMember)
				groups.DELETE("/:id/members/:userId", r.groupHandler.RemoveMember)
				groups.PUT("/:id/members/:userId/role", r.groupHandler.UpdateMemberRole)
				groups.POST("/:id/leave", r.groupHandler.LeaveGroup)
			}

			// Channel routes (Day 6)
			channels := protected.Group("/channels")
			{
				channels.POST("", r.channelHandler.CreateChannel)
				channels.GET("/public", r.channelHandler.GetPublicChannels)
				channels.GET("/owned", r.channelHandler.GetUserChannels)
				channels.GET("/subscriptions", r.channelHandler.GetSubscriptions)
				channels.GET("/:username", r.channelHandler.GetChannel)
				channels.POST("/:username/subscribe", r.channelHandler.Subscribe)
				channels.POST("/:username/unsubscribe", r.channelHandler.Unsubscribe)
				channels.DELETE("/:id", r.channelHandler.DeleteChannel)
			}

			// Media routes (Day 7-8)
			media := protected.Group("/media")
			{
				media.POST("/upload", r.mediaHandler.UploadMedia)
				media.POST("/voice", r.mediaHandler.UploadVoice)     // Day 8: Voice messages
				media.POST("/file", r.mediaHandler.UploadFile)       // Day 8: File upload
				media.GET("/storage", r.mediaHandler.GetStorageInfo) // Day 8: Storage info
				media.GET("", r.mediaHandler.GetUserMedia)
				media.GET("/:id", r.mediaHandler.GetMedia)
				media.DELETE("/:id", r.mediaHandler.DeleteMedia)
			}

			// Wallet routes (Day 9)
			wallet := protected.Group("/wallet")
			{
				wallet.POST("", r.walletHandler.AddWallet)
				wallet.GET("", r.walletHandler.GetWallets)
				wallet.GET("/transactions", r.walletHandler.GetTransactionHistory)
				wallet.GET("/transactions/by-signature/:signature", r.walletHandler.GetTransactionBySignature)
				wallet.GET("/by-address/:address", r.walletHandler.GetWalletByAddress)
				wallet.GET("/:id", r.walletHandler.GetWallet)
				wallet.POST("/:id/refresh", r.walletHandler.RefreshBalance)
				wallet.POST("/:id/default", r.walletHandler.SetDefault)
				wallet.POST("/:id/sync", r.walletHandler.SyncTransactions)
				wallet.POST("/:id/airdrop", r.walletHandler.RequestAirdrop)
				wallet.DELETE("/:id", r.walletHandler.DeleteWallet)
			}

			// Payment routes (Day 10)
			payments := protected.Group("/payments")
			{
				payments.POST("/request", r.paymentHandler.CreatePaymentRequest)
				payments.POST("/send", r.paymentHandler.SendPayment)
				payments.GET("/history", r.paymentHandler.GetPaymentHistory)
				payments.GET("/pending", r.paymentHandler.GetPendingPayments)
				payments.GET("/:id", r.paymentHandler.GetPaymentRequest)
				payments.POST("/:id/accept", r.paymentHandler.AcceptPaymentRequest)
				payments.POST("/:id/reject", r.paymentHandler.RejectPaymentRequest)
				payments.POST("/:id/cancel", r.paymentHandler.CancelPaymentRequest)
				payments.POST("/:id/confirm", r.paymentHandler.ConfirmPayment)
			}

			// Privacy & Security routes (Day 12)
			privacy := protected.Group("/privacy")
			{
				// Privacy settings
				privacy.GET("/settings", r.privacyHandler.GetPrivacySettings)
				privacy.PUT("/settings", r.privacyHandler.UpdatePrivacySettings)

				// User blocking
				privacy.POST("/block", r.privacyHandler.BlockUser)
				privacy.DELETE("/block/:userId", r.privacyHandler.UnblockUser)
				privacy.GET("/blocked", r.privacyHandler.GetBlockedUsers)

				// Disappearing messages
				privacy.POST("/disappearing", r.privacyHandler.SetDisappearingMessages)
				privacy.GET("/disappearing/:conversationId", r.privacyHandler.GetDisappearingMessagesConfig)
				privacy.DELETE("/disappearing/:conversationId", r.privacyHandler.DisableDisappearingMessages)

				// Two-Factor Authentication
				privacy.POST("/2fa/setup", r.privacyHandler.SetupTwoFactorAuth)
				privacy.POST("/2fa/enable", r.privacyHandler.EnableTwoFactorAuth)
				privacy.POST("/2fa/disable", r.privacyHandler.DisableTwoFactorAuth)
				privacy.GET("/2fa/status", r.privacyHandler.GetTwoFactorStatus)
			}

			// Day 13: Advanced Messaging Features
			// Message reactions, pinned messages, forwarding, search
			protected.POST("/messages/:id/react", r.messageHandler.AddReaction)
			protected.DELETE("/messages/:id/react", r.messageHandler.RemoveReaction)
			protected.GET("/messages/:id/reactions", r.messageHandler.GetMessageReactions)
			protected.POST("/messages/:id/pin", r.messageHandler.PinMessage)
			protected.DELETE("/messages/:id/pin", r.messageHandler.UnpinMessage)
			protected.GET("/conversations/:id/pinned", r.messageHandler.GetPinnedMessages)
			protected.POST("/messages/:id/forward", r.messageHandler.ForwardMessage)
			protected.POST("/messages/search", r.messageHandler.SearchMessages)

			// Status/Stories routes (Day 13)
			statuses := protected.Group("/statuses")
			{
				statuses.POST("", r.statusHandler.CreateStatus)
				statuses.GET("/feed", r.statusHandler.GetStatusFeed)
				statuses.GET("/user/:userId", r.statusHandler.GetUserStatuses)
				statuses.POST("/:id/view", r.statusHandler.ViewStatus)
				statuses.GET("/:id/views", r.statusHandler.GetStatusViews)
				statuses.DELETE("/:id", r.statusHandler.DeleteStatus)
			}

			// Contact routes (Day 13)
			contacts := protected.Group("/contacts")
			{
				contacts.POST("", r.contactHandler.AddContact)
				contacts.GET("", r.contactHandler.GetContacts)
				contacts.GET("/favorites", r.contactHandler.GetFavoriteContacts)
				contacts.GET("/search", r.contactHandler.SearchContacts)
				contacts.DELETE("/:contactId", r.contactHandler.RemoveContact)
				contacts.PUT("/:contactId/favorite", r.contactHandler.SetFavorite)

				// Contact invitations
				contacts.POST("/invites", r.contactHandler.SendInvite)
				contacts.GET("/invites/pending", r.contactHandler.GetPendingInvites)
				contacts.POST("/invites/:inviteId/accept", r.contactHandler.AcceptInvite)
				contacts.POST("/invites/:inviteId/reject", r.contactHandler.RejectInvite)
			}

			// Referral routes
			referrals := protected.Group("/referrals")
			{
				referrals.GET("/my-code", r.referralHandler.GetMyReferralCode)
				referrals.GET("/stats", r.referralHandler.GetMyReferralStats)
				referrals.GET("/history", r.referralHandler.GetMyReferralHistory)
				referrals.POST("/validate", r.referralHandler.ValidateReferralCode)
			}

			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", r.userHandler.GetProfile)
				users.PUT("/profile", r.userHandler.UpdateProfile)
				users.GET("/preferences", r.userHandler.GetPreferences)
				users.PUT("/preferences", r.userHandler.UpdatePreferences)
				users.POST("/check-wallet", r.userHandler.CheckWalletExists)
				users.GET("/:id", r.userHandler.GetUserByID)
			}

			// Invitation routes
			invitations := protected.Group("/invitations")
			{
				invitations.POST("", r.userHandler.SendInvitation)
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", r.notificationHandler.GetNotifications)
				notifications.GET("/unread", r.notificationHandler.GetUnreadCount)
				notifications.POST("/:id/read", r.notificationHandler.MarkAsRead)
				notifications.POST("/read-all", r.notificationHandler.MarkAllAsRead)
				notifications.DELETE("/:id", r.notificationHandler.DeleteNotification)

				// Notification settings
				notifications.GET("/settings", r.notificationHandler.GetSettings)
				notifications.PUT("/settings", r.notificationHandler.UpdateSettings)
			}
		}
	}

	logger.Info("Router setup completed",
		zap.String("mode", mode),
	)

	return router
}

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
