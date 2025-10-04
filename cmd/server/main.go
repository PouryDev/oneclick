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
	"github.com/PouryDev/oneclick/internal/app/crypto"
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
	clusterRepo := repo.NewClusterRepository(db)
	repositoryRepo := repo.NewRepositoryRepository(db)
	appRepo := repo.NewApplicationRepository(db)
	releaseRepo := repo.NewReleaseRepository(db)
	gitServerRepo := repo.NewGitServerRepository(db)
	runnerRepo := repo.NewRunnerRepository(db)
	jobRepo := repo.NewJobRepository(db)
	domainRepo := repo.NewDomainRepository(db)

	// Initialize crypto
	cryptoService, err := crypto.NewCrypto()
	if err != nil {
		logger.Fatal("Failed to initialize crypto service", zap.Error(err))
	}

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret)
	orgService := services.NewOrganizationService(orgRepo, userRepo)
	clusterService := services.NewClusterService(clusterRepo, orgRepo, cryptoService)
	repositoryService := services.NewRepositoryService(repositoryRepo, orgRepo, cryptoService)
	applicationService := services.NewApplicationService(appRepo, releaseRepo, clusterRepo, repositoryRepo, orgRepo)
	gitServerService := services.NewGitServerService(gitServerRepo, jobRepo, orgRepo, cryptoService, logger)
	runnerService := services.NewRunnerService(runnerRepo, jobRepo, orgRepo, cryptoService, logger)
	jobService := services.NewJobService(jobRepo, orgRepo, logger)
	domainService := services.NewDomainService(domainRepo, appRepo, jobRepo, orgRepo, cryptoService, logger)
	// For now, we'll pass nil for the Kubernetes client
	// In a real implementation, you would create a Kubernetes client factory
	// that creates clients per request based on the cluster's kubeconfig
	podService := services.NewPodService(appRepo, clusterRepo, orgRepo, cryptoService, nil, logger)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	orgHandler := handlers.NewOrganizationHandler(orgService)
	clusterHandler := handlers.NewClusterHandler(clusterService)
	repositoryHandler := handlers.NewRepositoryHandler(repositoryService)
	webhookHandler := handlers.NewWebhookHandler(repositoryService, logger)
	applicationHandler := handlers.NewApplicationHandler(applicationService)
	gitServerHandler := handlers.NewGitServerHandler(gitServerService, logger)
	runnerHandler := handlers.NewRunnerHandler(runnerService, logger)
	jobHandler := handlers.NewJobHandler(jobService, logger)
	domainHandler := handlers.NewDomainHandler(domainService, logger)
	podHandler := handlers.NewPodHandler(podService, logger)

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

			// Cluster management routes
			clusters := orgSpecific.Group("/clusters")
			clusters.Use(middleware.RequireMemberMiddleware())
			{
				clusters.POST("", clusterHandler.CreateCluster)
				clusters.GET("", clusterHandler.GetClustersByOrg)
				clusters.POST("/import", clusterHandler.ImportCluster)
			}

			// Repository management routes
			repos := orgSpecific.Group("/repos")
			repos.Use(middleware.RequireMemberMiddleware())
			{
				repos.POST("", repositoryHandler.CreateRepository)
				repos.GET("", repositoryHandler.GetRepositoriesByOrg)
			}

			// Git server management routes
			gitservers := orgSpecific.Group("/gitservers")
			gitservers.Use(middleware.RequireAdminOrOwnerMiddleware())
			{
				gitservers.POST("", gitServerHandler.CreateGitServer)
				gitservers.GET("", gitServerHandler.GetGitServersByOrg)
			}

			// Runner management routes
			runners := orgSpecific.Group("/runners")
			runners.Use(middleware.RequireAdminOrOwnerMiddleware())
			{
				runners.POST("", runnerHandler.CreateRunner)
				runners.GET("", runnerHandler.GetRunnersByOrg)
			}

			// Job management routes
			jobs := orgSpecific.Group("/jobs")
			jobs.Use(middleware.RequireMemberMiddleware())
			{
				jobs.GET("", jobHandler.GetJobsByOrg)
			}
		}
	}

	// Global cluster routes (require authentication)
	clusters := router.Group("/clusters")
	clusters.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		clusters.GET("/:clusterId", clusterHandler.GetCluster)
		clusters.GET("/:clusterId/status", clusterHandler.GetClusterHealth)
		clusters.DELETE("/:clusterId", middleware.RequireAdminOrOwnerMiddleware(), clusterHandler.DeleteCluster)

		// Application management routes within clusters
		clusters.POST("/:clusterId/apps", applicationHandler.CreateApplication)
		clusters.GET("/:clusterId/apps", applicationHandler.GetApplicationsByCluster)
	}

	// Global repository routes (require authentication)
	repos := router.Group("/repos")
	repos.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		repos.GET("/:repoId", repositoryHandler.GetRepository)
		repos.DELETE("/:repoId", middleware.RequireAdminOrOwnerMiddleware(), repositoryHandler.DeleteRepository)
	}

	// Global application routes (require authentication)
	apps := router.Group("/apps")
	apps.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		apps.GET("/:appId", applicationHandler.GetApplication)
		apps.DELETE("/:appId", middleware.RequireAdminOrOwnerMiddleware(), applicationHandler.DeleteApplication)
		apps.POST("/:appId/deploy", applicationHandler.DeployApplication)
		apps.GET("/:appId/releases", applicationHandler.GetReleasesByApplication)
		apps.POST("/:appId/releases/:releaseId/rollback", applicationHandler.RollbackApplication)

		// Domain management routes
		apps.POST("/:appId/domains", domainHandler.CreateDomain)
		apps.GET("/:appId/domains", domainHandler.GetDomainsByApp)

		// Pod management routes
		apps.GET("/:appId/pods", podHandler.GetPodsByApp)
	}

	// Global git server routes (require authentication)
	gitservers := router.Group("/gitservers")
	gitservers.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		gitservers.GET("/:gitServerId", gitServerHandler.GetGitServer)
		gitservers.DELETE("/:gitServerId", middleware.RequireAdminOrOwnerMiddleware(), gitServerHandler.DeleteGitServer)
	}

	// Global runner routes (require authentication)
	runners := router.Group("/runners")
	runners.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		runners.GET("/:runnerId", runnerHandler.GetRunner)
		runners.DELETE("/:runnerId", middleware.RequireAdminOrOwnerMiddleware(), runnerHandler.DeleteRunner)
	}

	// Global domain routes (require authentication)
	domains := router.Group("/domains")
	domains.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		domains.GET("/:domainId", domainHandler.GetDomain)
		domains.DELETE("/:domainId", middleware.RequireAdminOrOwnerMiddleware(), domainHandler.DeleteDomain)
		domains.POST("/:domainId/certificates", domainHandler.RequestCertificate)
		domains.GET("/:domainId/certificates", domainHandler.GetCertificateStatus)
	}

	// Global pod routes (require authentication)
	pods := router.Group("/pods")
	pods.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		pods.GET("/:podId", podHandler.GetPodDetail)
		pods.GET("/:podId/logs", podHandler.GetPodLogs)
		pods.GET("/:podId/describe", podHandler.GetPodDescribe)
		pods.POST("/:podId/terminal", podHandler.ExecInPod)
	}

	// Public webhook routes (no authentication required)
	webhooks := router.Group("/hooks")
	{
		webhooks.POST("/git", webhookHandler.GitWebhook)
		webhooks.GET("/test", webhookHandler.TestWebhook)
	}

	// Start server
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Starting server", zap.String("port", port))

	if err := router.Run(port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
