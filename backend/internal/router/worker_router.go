package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/vikukumar/pushpaka/internal/config"
	"github.com/vikukumar/pushpaka/internal/handlers"
	"github.com/vikukumar/pushpaka/internal/middleware"
)

// NewWorkerRouter builds the Gin engine for the isolated Worker Management API.
func NewWorkerRouter(cfg *config.Config, workerHandler *handlers.WorkerHandler) *gin.Engine {
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
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	worker := r.Group("/api/v1/worker")
	{
		worker.POST("/register", workerHandler.Register)
		worker.GET("/ws", workerHandler.ConnectWS)
		worker.GET("/poll", workerHandler.Poll)
	}

	return r
}
