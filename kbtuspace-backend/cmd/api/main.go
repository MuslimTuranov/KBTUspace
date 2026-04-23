package main

import (
	"context"
	"log/slog"
	"time"

	"kbtuspace-backend/internal/auth"
	"kbtuspace-backend/internal/events"
	"kbtuspace-backend/internal/middleware"
	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/internal/posts"
	"kbtuspace-backend/internal/reports"
	"kbtuspace-backend/internal/users"
	"kbtuspace-backend/pkg/cache"
	"kbtuspace-backend/pkg/config"
	"kbtuspace-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "kbtuspace-backend/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           UniHub API
// @version         1.0
// @description     This is the central API for the UniHub university ecosystem.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Type "Bearer " followed by a space and JWT token.
func main() {
	_ = godotenv.Load()

	logger.Init()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", slog.Any("error", err))
		return
	}

	ctx := context.Background()

	db, err := models.InitDB(cfg)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to database", slog.Any("error", err))
		return
	}
	defer db.Close()

	slog.InfoContext(ctx, "Connected to database")

	cacheClient, err := cache.NewRedisCache(cfg.RedisURL, 10*time.Minute)
	if err != nil {
		slog.InfoContext(ctx, "Redis cache disabled", slog.Any("error", err))
		cacheClient = nil
	} else {
		slog.InfoContext(ctx, "Redis cache initialized")
	}

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, []byte(cfg.JWTSecret))
	authHandler := auth.NewHandler(authService)

	userRepo := users.NewRepository(db)
	userService := users.NewService(userRepo)
	userHandler := users.NewHandler(userService)

	postRepo := posts.NewRepository(db)
	postService := posts.NewService(postRepo, cacheClient)
	postHandler := posts.NewHandler(postService)

	eventRepo := events.NewRepository(db)
	eventService := events.NewService(eventRepo, cacheClient)
	eventHandler := events.NewHandler(eventService)

	reportRepo := reports.NewRepository(db)
	reportService := reports.NewService(reportRepo)
	reportHandler := reports.NewHandler(reportService)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	// @Summary     Health check
	// @Description Returns pong if API is running
	// @Tags        system
	// @Produce     json
	// @Success     200 {object} map[string]interface{}
	// @Router      /ping [get]
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "UniHub API is running",
		})
	})

	api := router.Group("/api/v1")
	{
		// Auth routes (public)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.RequireAuth([]byte(cfg.JWTSecret)))
		{
			// Profile
			protected.GET("/profile", userHandler.GetProfile)
			protected.PUT("/profile", userHandler.UpdateProfile)

			// Posts
			protected.POST("/posts", postHandler.Create)
			protected.GET("/posts", postHandler.GetAll)
			protected.GET("/posts/:id", postHandler.GetByID)
			protected.PUT("/posts/:id", postHandler.Update)
			protected.DELETE("/posts/:id", postHandler.Delete)
			protected.PATCH("/posts/:id/pin", postHandler.Pin)

			// Events
			protected.GET("/events", eventHandler.GetAll)
			protected.GET("/events/:id", eventHandler.GetByID)
			protected.POST("/events/:id/register", eventHandler.Register)
			protected.POST("/reports", reportHandler.Create)

			// Organizer-only routes
			organizerOnly := protected.Group("/events")
			organizerOnly.Use(middleware.RequireRole("organizer", "admin"))
			{
				organizerOnly.POST("/", eventHandler.Create)
				organizerOnly.PUT("/:id", eventHandler.Update)
				organizerOnly.DELETE("/:id", eventHandler.Delete)
			}

			adminOnly := protected.Group("/admin")
			adminOnly.Use(middleware.RequireRole("admin"))
			{
				adminOnly.GET("/moderation/global-content", func(c *gin.Context) {
					contentType := c.DefaultQuery("type", "all")
					switch contentType {
					case "posts":
						postHandler.ListPendingGlobal(c)
					case "events":
						eventHandler.ListPendingGlobal(c)
					default:
						posts, postsErr := postService.ListPendingGlobal()
						events, eventsErr := eventService.ListPendingGlobal()
						if postsErr != nil || eventsErr != nil {
							c.JSON(500, gin.H{"error": "Failed to fetch pending global content"})
							return
						}
						if posts == nil {
							posts = []models.Post{}
						}
						if events == nil {
							events = []models.Post{}
						}
						c.JSON(200, gin.H{
							"posts":  posts,
							"events": events,
						})
					}
				})

				adminOnly.GET("/reports", reportHandler.List)
				adminOnly.PATCH("/reports/:id/close", reportHandler.Close)

				adminOnly.PATCH("/posts/:id/approve", postHandler.Approve)
				adminOnly.PATCH("/posts/:id/reject", postHandler.Reject)
				adminOnly.DELETE("/posts/:id", postHandler.AdminDelete)

				adminOnly.PATCH("/events/:id/approve", eventHandler.Approve)
				adminOnly.PATCH("/events/:id/reject", eventHandler.Reject)
				adminOnly.DELETE("/events/:id", eventHandler.AdminDelete)

				adminOnly.PATCH("/users/:id", userHandler.AdminUpdateUser)
			}
		}
	}

	addr := ":" + cfg.Port
	slog.InfoContext(ctx, "Starting server", slog.String("address", addr))
	if err := router.Run(addr); err != nil {
		slog.ErrorContext(ctx, "Failed to start server", slog.Any("error", err))
	}
}
