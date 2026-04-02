package main

import (
	"log"
	"os"
	"time"

	"kbtuspace-backend/internal/auth"
	"kbtuspace-backend/internal/events"
	"kbtuspace-backend/internal/middleware"
	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/internal/posts"
	"kbtuspace-backend/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := models.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to Database!")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.Default()

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)

	postRepo := posts.NewRepository(db)
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	postCache, err := cache.NewRedisCache(redisURL, 10*time.Minute)
	if err != nil {
		log.Printf("Redis cache disabled: %v", err)
	}

	postService := posts.NewService(postRepo, postCache)
	postHandler := posts.NewHandler(postService)

	eventRepo := events.NewRepository(db)
	eventService := events.NewService(eventRepo, postCache)
	eventHandler := events.NewHandler(eventService)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong", "status": "UniHub API is running!"})
	})

	api := router.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		protected := api.Group("/")
		protected.Use(middleware.RequireAuth())
		{
			protected.GET("/profile", func(c *gin.Context) {
				userID, _ := c.Get("userID")
				role, _ := c.Get("role")

				c.JSON(200, gin.H{
					"message": "Welcome to your profile",
					"user_id": userID,
					"role":    role,
				})
			})

			protected.POST("/posts", postHandler.Create)
			protected.GET("/posts", postHandler.GetAll)
			protected.GET("/posts/:id", postHandler.GetByID)
			protected.PUT("/posts/:id", postHandler.Update)
			protected.DELETE("/posts/:id", postHandler.Delete)

			protected.GET("/events", eventHandler.GetAll)
			protected.GET("/events/:id", eventHandler.GetByID)
			protected.POST("/events/:id/register", eventHandler.Register)

			organizerOnly := protected.Group("/events")
			organizerOnly.Use(middleware.RequireRole("organizer", "admin"))
			{
				organizerOnly.POST("/", eventHandler.Create)
				organizerOnly.PUT("/:id", eventHandler.Update)
				organizerOnly.DELETE("/:id", eventHandler.Delete)
			}
		}
	}

	log.Printf("Starting server on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
