package routes

import (
	"nexfi-backend/api/handlers"
	"nexfi-backend/api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Initialize handlers
	oauthHandler := handlers.NewOAuthHandler()
	healthCheckHandler := handlers.NewHealthCheckHandler()
	userHandler := handlers.NewUserHandler(nil)
	projectHandler := handlers.NewProjectHandler(db)
	componentHandler := handlers.NewComponentHandler(db)
	challengeHandler := handlers.NewChallengeHandler(db)
	notificationHandler := handlers.NewNotificationHandler(db)
	gamificationHandler := handlers.NewGamificationHandler(db)
	docHandler := handlers.NewDocHandler(db)

	// Apply CORS middleware globally
	router.Use(middleware.CORSMiddleware())

	// Serve uploaded files statically
	router.Static("/uploads", "./uploads")

	// Base API group
	apiV1 := router.Group("/api/v1")
	{
		// Health check route
		apiV1.GET("/health", healthCheckHandler.HealthCheck)

		// Base route
		apiV1.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Welcome to NexFlux Virtual Lab API",
				"version": "1.0.0",
				"status":  "running",
			})
		})

		// ========================================
		// AUTH ROUTES (Public)
		// ========================================
		auth := apiV1.Group("/auth")
		{
			// Email/Password authentication
			auth.POST("/register", oauthHandler.RegisterWithPassword)
			auth.POST("/login", oauthHandler.LoginWithPassword)

			// OAuth providers list
			auth.GET("/providers", oauthHandler.GetSupportedProviders)

			// Google OAuth
			auth.GET("/google", oauthHandler.GetOAuthURL)
			auth.GET("/google/callback", oauthHandler.HandleOAuthCallback)
			auth.POST("/google/callback", oauthHandler.HandleOAuthCallbackPost)

			// GitHub OAuth
			auth.GET("/github", oauthHandler.GetOAuthURL)
			auth.GET("/github/callback", oauthHandler.HandleOAuthCallback)
			auth.POST("/github/callback", oauthHandler.HandleOAuthCallbackPost)

			// Apple OAuth (uses form_post for callback)
			auth.GET("/apple", oauthHandler.GetOAuthURL)
			auth.GET("/apple/callback", oauthHandler.HandleOAuthCallback)
			auth.POST("/apple/callback", oauthHandler.HandleOAuthCallbackPost)

			// Generic OAuth endpoint (supports all providers via path param)
			auth.GET("/:provider", oauthHandler.GetOAuthURL)
			auth.GET("/:provider/callback", oauthHandler.HandleOAuthCallback)
			auth.POST("/:provider/callback", oauthHandler.HandleOAuthCallbackPost)
		}

		// ========================================
		// PUBLIC ROUTES (No auth required)
		// ========================================

		// Component categories (public)
		apiV1.GET("/components/categories", componentHandler.GetCategories)

		// ======== DOCUMENTATION ROUTES (Public) ========
		docs := apiV1.Group("/docs")
		{
			// Categories
			docs.GET("/categories", docHandler.GetCategories)
			docs.GET("/categories/:slug", docHandler.GetCategoryBySlug)

			// Articles
			docs.GET("/articles", docHandler.ListArticles)
			docs.GET("/articles/popular", docHandler.GetPopularArticles)
			docs.GET("/articles/featured", docHandler.GetFeaturedArticles)
			docs.GET("/articles/:slug", docHandler.GetArticleBySlug)
			docs.POST("/articles/:slug/view", docHandler.IncrementArticleView)

			// Videos
			docs.GET("/videos", docHandler.ListVideos)
			docs.GET("/videos/:id", docHandler.GetVideoByID)

			// Search
			docs.GET("/search", docHandler.Search)
		}

		// ========================================
		// PROTECTED ROUTES (Requires JWT)
		// ========================================
		protected := apiV1.Group("/")
		protected.Use(middleware.JWTAuthMiddleware())
		{
			// ======== USER ROUTES ========
			user := protected.Group("/users")
			{
				user.GET("/me", userHandler.GetUserProfile)
				user.PUT("/me", userHandler.UpdateUserProfile)
				user.GET("/me/stats", gamificationHandler.GetUserStats)
				user.POST("/me/avatar", userHandler.UploadAvatar)
				user.DELETE("/me/avatar", userHandler.DeleteAvatar)
				user.GET("/me/settings", userHandler.GetUserSettings)
				user.PUT("/me/settings", userHandler.UpdateUserSettings)
			}

			// ======== PROJECT ROUTES ========
			projects := protected.Group("/projects")
			{
				projects.GET("", projectHandler.ListProjects)
				projects.POST("", projectHandler.CreateProject)
				projects.GET("/templates", projectHandler.GetTemplates)
				projects.GET("/:id", projectHandler.GetProject)
				projects.PUT("/:id", projectHandler.UpdateProject)
				projects.DELETE("/:id", projectHandler.DeleteProject)
				projects.POST("/:id/duplicate", projectHandler.DuplicateProject)
				projects.PUT("/:id/favorite", projectHandler.ToggleFavorite)
				projects.GET("/:id/collaborators", projectHandler.GetCollaborators)
			}

			// ======== COMPONENT ROUTES ========
			components := protected.Group("/components")
			{
				components.GET("", componentHandler.ListComponents)
				components.GET("/search", componentHandler.SearchComponents)
				components.GET("/favorites", componentHandler.GetFavorites)
				components.GET("/requests", componentHandler.GetUserRequests)
				components.POST("/request", componentHandler.CreateRequest)
				components.GET("/:id", componentHandler.GetComponent)
				components.POST("/:id/favorite", componentHandler.ToggleFavorite)
			}

			// ======== CHALLENGE ROUTES ========
			challenges := protected.Group("/challenges")
			{
				challenges.GET("", challengeHandler.ListChallenges)
				challenges.GET("/daily", challengeHandler.GetDailyChallenge)
				challenges.GET("/progress", challengeHandler.GetUserProgress)
				challenges.GET("/:id", challengeHandler.GetChallenge)
				challenges.POST("/:id/start", challengeHandler.StartChallenge)
				challenges.PUT("/:id/progress", challengeHandler.UpdateProgress)
				challenges.POST("/:id/submit", challengeHandler.SubmitChallenge)
			}

			// ======== NOTIFICATION ROUTES ========
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", notificationHandler.ListNotifications)
				notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
				notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
				notifications.DELETE("/:id", notificationHandler.DeleteNotification)
			}

			// ======== GAMIFICATION ROUTES ========
			protected.GET("/achievements", gamificationHandler.GetAllAchievements)
			protected.GET("/achievements/user", gamificationHandler.GetUserAchievements)
			protected.GET("/leaderboard", gamificationHandler.GetLeaderboard)
			protected.GET("/streak", gamificationHandler.GetStreak)
		}
	}
}
