package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/api/handlers"
	"github.com/PouryDev/oneclick/internal/api/middleware"
	"github.com/PouryDev/oneclick/internal/app/services"
	"github.com/PouryDev/oneclick/internal/config"
	"github.com/PouryDev/oneclick/internal/repo"

	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Connect to database
	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	logger.Info("Successfully connected to database")

	// Initialize repositories
	userRepo := repo.NewUserRepository(db)
	orgRepo := repo.NewOrganizationRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret)
	orgService := services.NewOrganizationService(orgRepo, userRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	orgHandler := handlers.NewOrganizationHandler(orgService)

	// Setup Gin router
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", middleware.AuthMiddleware(cfg.JWT.Secret), authHandler.Me)
	}

	// Organization routes
	orgs := router.Group("/orgs")
	orgs.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		orgs.POST("", orgHandler.CreateOrganization)
		orgs.GET("", orgHandler.GetUserOrganizations)

		// Organization-specific routes with access control
		orgSpecific := orgs.Group("/:orgId")
		orgSpecific.Use(middleware.OrganizationAccessMiddleware(orgRepo))
		{
			orgSpecific.GET("", orgHandler.GetOrganization)
			orgSpecific.DELETE("", middleware.RequireOwnerMiddleware(), orgHandler.DeleteOrganization)

			// Member management routes
			members := orgSpecific.Group("/members")
			members.Use(middleware.RequireAdminOrOwnerMiddleware())
			{
				members.POST("", orgHandler.AddMember)
				members.PATCH("/:userId", orgHandler.UpdateMemberRole)
				members.DELETE("/:userId", orgHandler.RemoveMember)
			}
		}
	}

	// Start server
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Starting server", zap.String("port", port))

	if err := router.Run(port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
