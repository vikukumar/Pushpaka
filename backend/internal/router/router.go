package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/vikukumar/Pushpaka/internal/config"
	"github.com/vikukumar/Pushpaka/internal/handlers"
	"github.com/vikukumar/Pushpaka/internal/middleware"
	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/internal/services"
)

func New(cfg *config.Config, db *sqlx.DB, rdb *redis.Client) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.SecureHeaders())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-API-Key"},
		AllowCredentials: true,
	}))

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	projectRepo := repositories.NewProjectRepository(db)
	deploymentRepo := repositories.NewDeploymentRepository(db)
	logRepo := repositories.NewLogRepository(db)
	domainRepo := repositories.NewDomainRepository(db)
	envRepo := repositories.NewEnvVarRepository(db)

	// Services
	authSvc := services.NewAuthService(userRepo, cfg)
	projectSvc := services.NewProjectService(projectRepo)
	deploymentSvc := services.NewDeploymentService(deploymentRepo, projectRepo, envRepo, rdb)
	logSvc := services.NewLogService(logRepo)
	domainSvc := services.NewDomainService(domainRepo, projectRepo)
	envSvc := services.NewEnvService(envRepo, projectRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authSvc)
	projectHandler := handlers.NewProjectHandler(projectSvc)
	deploymentHandler := handlers.NewDeploymentHandler(deploymentSvc)
	logHandler := handlers.NewLogHandler(logSvc)
	domainHandler := handlers.NewDomainHandler(domainSvc)
	envHandler := handlers.NewEnvHandler(envSvc)
	healthHandler := handlers.NewHealthHandler(db, rdb)

	// Auth middleware
	authMW := middleware.JWT(authSvc)

	// API v1
	api := r.Group("/api/v1")
	{
		// Health & Metrics (public)
		api.GET("/health", healthHandler.Health)
		api.GET("/ready", healthHandler.Ready)
		api.GET("/metrics", handlers.MetricsHandler())

		// Auth (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(authMW)
		{
			// Projects
			projects := protected.Group("/projects")
			{
				projects.POST("", projectHandler.Create)
				projects.GET("", projectHandler.List)
				projects.GET("/:id", projectHandler.Get)
				projects.DELETE("/:id", projectHandler.Delete)
			}

			// Deployments
			deployments := protected.Group("/deployments")
			{
				deployments.POST("", deploymentHandler.Deploy)
				deployments.GET("", deploymentHandler.List)
				deployments.GET("/:id", deploymentHandler.Get)
				deployments.POST("/:id/rollback", deploymentHandler.Rollback)
			}

			// Logs (REST + WebSocket)
			logs := protected.Group("/logs")
			{
				logs.GET("/:id", logHandler.GetLogs)
				logs.GET("/:id/stream", logHandler.StreamLogs)
			}

			// Domains
			domains := protected.Group("/domains")
			{
				domains.POST("", domainHandler.Add)
				domains.GET("", domainHandler.List)
				domains.DELETE("/:id", domainHandler.Delete)
			}

			// Environment Variables
			env := protected.Group("/env")
			{
				env.POST("", envHandler.Set)
				env.GET("", envHandler.List)
				env.DELETE("", envHandler.Delete)
			}
		}
	}

	return r
}
