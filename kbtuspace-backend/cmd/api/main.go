package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"kbtuspace-backend/internal/auth"
	"kbtuspace-backend/internal/events"
	"kbtuspace-backend/internal/middleware"
	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/internal/posts"
	"kbtuspace-backend/internal/users"
	"kbtuspace-backend/internal/worker"
	"kbtuspace-backend/pkg/cache"
	"kbtuspace-backend/pkg/config"
	"kbtuspace-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Initialize logger
	logger.Init()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", slog.Any("error", err))
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Connect to database
	db, err := models.InitDB(cfg)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to database", slog.Any("error", err))
		return
	}
	defer db.Close()

	slog.InfoContext(ctx, "Connected to database")

	if err := models.SeedDefaults(db); err != nil {
		slog.ErrorContext(ctx, "Failed to seed defaults", slog.Any("error", err))
		return
	}

	// Initialize Redis cache
	cacheClient, err := cache.NewRedisCache(cfg.RedisURL, 10*time.Minute)
	if err != nil {
		slog.InfoContext(ctx, "Redis cache disabled", slog.Any("error", err))
		cacheClient = nil
	} else {
		slog.InfoContext(ctx, "Redis cache initialized")
	}

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Initialize services and handlers
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

	// Health check endpoint
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "UniHub API is running",
		})
	})

	// API routes
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
			protected.DELETE("/events/:id/register", eventHandler.CancelRegistration)

			// Organizer-only routes
			organizerOnly := protected.Group("/events")
			organizerOnly.Use(middleware.RequireRole("organizer", "admin"))
			{
				organizerOnly.POST("/", eventHandler.Create)
				organizerOnly.PUT("/:id", eventHandler.Update)
				organizerOnly.DELETE("/:id", eventHandler.Delete)
				organizerOnly.PATCH("/:id/attendance/:userId", eventHandler.MarkAttendance)
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

	// Start server
	reminderWorker := worker.NewReminderWorker(db)
	go reminderWorker.Start(ctx)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		slog.InfoContext(ctx, "Starting server", slog.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "Server error", slog.Any("error", err))
		}
	}()

	<-ctx.Done()
	slog.InfoContext(context.Background(), "Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, "Forced shutdown", slog.Any("error", err))
	}

	slog.InfoContext(context.Background(), "Server stopped")
}
