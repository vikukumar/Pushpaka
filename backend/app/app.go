// Package app exposes the Pushpaka API server as a callable function,
// allowing it to be embedded in the combined pushpaka binary.
package app

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/internal/config"
	"github.com/vikukumar/Pushpaka/internal/handlers"
	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/internal/router"
	"github.com/vikukumar/Pushpaka/internal/services"
	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/database"
	"github.com/vikukumar/Pushpaka/pkg/models"
	"github.com/vikukumar/Pushpaka/pkg/tunnel"
	"github.com/vikukumar/Pushpaka/queue"
	"github.com/vikukumar/Pushpaka/ui"
)

// OpenDB opens the database using the given driver and DSN.
// Exposed so that callers embedding multiple components (e.g. the combined
// binary) can open ONE shared connection pool and pass it to both the API
// and the worker, preventing SQLite BUSY_SNAPSHOT errors in WAL mode.
func OpenDB(driver, dsn string) (*gorm.DB, error) {
	return basemodel.Connect(driver, dsn, "development")
}

// RunOptions configures optional behaviour for Run.
type RunOptions struct {
	// InProcessQueue, when non-nil, is used instead of Redis for deployment jobs.
	// Intended for dev mode where the embedded worker reads from the same queue.
	InProcessQueue *queue.InProcess

	// DB, when non-nil, is used instead of opening a new database connection.
	// The caller is responsible for closing the DB after RunWithOptions returns.
	// Used in all-in-one mode to share one SQLite pool between API and worker.
	DB *gorm.DB
}

// Run starts the Pushpaka API server with default options and blocks until ctx is cancelled.
func Run(ctx context.Context) error {
	return RunWithOptions(ctx, RunOptions{})
}

// RunWithOptions starts the Pushpaka API server with the supplied options.
func RunWithOptions(ctx context.Context, opts RunOptions) error {
	cfg := config.Load()

	var db *gorm.DB
	if opts.DB != nil {
		// Caller already opened a shared pool; we must not close it here.
		db = opts.DB
	} else {
		var err error
		db, err = basemodel.Connect(cfg.DatabaseDriver, cfg.DatabaseURL, cfg.AppEnv)
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
		defer func() {
			sqlDB, err := db.DB()
			if err == nil {
				sqlDB.Close()
			}
		}()
	}

	// Background Initialization: AutoMigrate and System Config
	go func() {
		log.Info().Msg("starting background database migration...")
		err := db.AutoMigrate(
			&models.User{},
			&models.Project{},
			&models.Domain{},
			&models.EnvVar{},
			&models.AuditLog{},
			&models.Deployment{},
			&models.DeploymentAction{},
			&models.DeploymentBackup{},
			&models.DeploymentCodeSignature{},
			&models.DeploymentInstance{},
			&models.DeploymentLog{},
			&models.DeploymentSyncHistory{},
			&models.GitAutoSyncConfig{},
			&models.GitChange{},
			&models.GitSyncTrack{},
			&models.NotificationConfig{},
			&models.WebhookConfig{},
			&models.AIConfig{},
			&models.RAGDocument{},
			&models.AIMonitorAlert{},
			&models.AITokenUsage{},
			&models.K8sConfig{},
			&models.SystemConfig{},
			&models.WorkerNode{},
		)
		if err != nil {
			log.Error().Err(err).Msg("database migration failed in background")
			// We don't exit here because the server is already starting/started.
			// Handlers will return 503 until DBReady is true.
			return
		}

		// Initialize System Configuration (ZoneID)
		systemCfgRepo := repositories.NewSystemConfigRepository(db)
		zoneID, err := systemCfgRepo.Get("ZONE_ID")
		if err != nil {
			zoneID = uuid.New().String()
			if setErr := systemCfgRepo.Set("ZONE_ID", zoneID); setErr != nil {
				log.Error().Err(setErr).Msg("failed to save new ZONE_ID")
			} else {
				log.Info().Str("zone_id", zoneID).Msg("generated new installation ZoneID")
			}
		} else {
			log.Info().Str("zone_id", zoneID).Msg("loaded installation ZoneID")
		}

		// Mark as ready
		basemodel.SetDBReady(true)
		log.Info().Msg("background database initialization complete")
	}()

	// Repositories for system-level services
	systemCfgRepo := repositories.NewSystemConfigRepository(db)
	workerNodeRepo := repositories.NewWorkerNodeRepository(db)

	// Redis is optional: skipped when REDIS_URL is empty or an in-process queue is used.
	var rdb *redis.Client
	if opts.InProcessQueue != nil {
		log.Info().Msg("using in-process job queue (dev mode)")
	} else if cfg.RedisURL != "" {
		var redisErr error
		rdb, redisErr = database.NewRedis(cfg.RedisURL)
		if redisErr != nil {
			log.Warn().Err(redisErr).Msg("redis unavailable - deployment triggers disabled")
		} else {
			defer rdb.Close()
		}
	} else {
		log.Warn().Msg("REDIS_URL not set and no in-process queue -- deployment triggers disabled")
	}

	// Repositories
	projectRepo := repositories.NewProjectRepository(db)
	deploymentRepo := repositories.NewDeploymentRepository(db)
	domainRepo := repositories.NewDomainRepository(db)
	envRepo := repositories.NewEnvVarRepository(db)
	logRepo := repositories.NewLogRepository(db)
	auditRepo := repositories.NewAuditRepository(db)
	notifRepo := repositories.NewNotificationRepository(db)
	webhookRepo := repositories.NewWebhookRepository(db)
	aiConfigRepo := repositories.NewAIConfigRepository(db)
	userRepo := repositories.NewUserRepository(db)

	// Services
	authSvc := services.NewAuthService(userRepo, cfg)
	projectSvc := services.NewProjectService(projectRepo)
	deploymentSvc := services.NewDeploymentService(deploymentRepo, projectRepo, envRepo, domainRepo, rdb, opts.InProcessQueue, cfg.BaseURL)
	logSvc := services.NewLogService(logRepo)
	domainSvc := services.NewDomainService(domainRepo, projectRepo)
	envSvc := services.NewEnvService(envRepo, projectRepo)
	auditSvc := services.NewAuditService(auditRepo)
	notifSvc := services.NewNotificationService(notifRepo, cfg)
	oauthSvc := services.NewOAuthService(userRepo, cfg, authSvc, db)
	webhookSvc := services.NewWebhookService(webhookRepo, projectRepo, deploymentSvc, cfg)
	aiSvc := services.NewAIService(cfg)
	aiExecutor := services.NewAIToolsExecutor(deploymentSvc, logSvc, aiSvc)

	reg := &router.ServiceRegistry{
		AuthSvc:        authSvc,
		ProjectSvc:     projectSvc,
		DeploymentSvc:  deploymentSvc,
		LogSvc:         logSvc,
		DomainSvc:      domainSvc,
		EnvSvc:         envSvc,
		AuditSvc:       auditSvc,
		NotifSvc:       notifSvc,
		OAuthSvc:       oauthSvc,
		WebhookSvc:     webhookSvc,
		AISvc:          aiSvc,
		WorkerSvc:      services.NewWorkerNodeService(workerNodeRepo, systemCfgRepo),
		AIExecutor:     aiExecutor,
		UserRepo:       userRepo,
		ProjectRepo:    projectRepo,
		DeploymentRepo: deploymentRepo,
		LogRepo:        logRepo,
		DomainRepo:     domainRepo,
		EnvRepo:        envRepo,
		AuditRepo:      auditRepo,
		NotifRepo:      notifRepo,
		WebhookRepo:    webhookRepo,
		AIConfigRepo:   aiConfigRepo,
		WorkerRepo:     workerNodeRepo,
		SystemRepo:     systemCfgRepo,
	}

	// Background Tasks
	go deploymentSvc.StartAutoSyncLoop(ctx)

	// Graceful Shutdown
	go func() {
		<-ctx.Done()
		deploymentSvc.Shutdown()
	}()

	// Detect whether the frontend was compiled into the binary.
	// In dev mode ui/dist only contains a placeholder, so uiFS stays nil.
	var uiFS fs.FS
	if _, ferr := ui.FS.Open("dist/index.html"); ferr == nil {
		if sub, serr := fs.Sub(ui.FS, "dist"); serr == nil {
			uiFS = sub
			log.Info().Msg("serving embedded frontend")
		}
	}

	// Main User API Router
	r := router.New(cfg, db, rdb, uiFS, opts.InProcessQueue, reg)

	// Worker Management Router
	workerNodeSvc := services.NewWorkerNodeService(workerNodeRepo, systemCfgRepo)
	workerHandler := handlers.NewWorkerHandler(workerNodeSvc)
	workerR := router.NewWorkerRouter(cfg, workerHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	workerSrv := &http.Server{
		Addr:         ":" + cfg.WorkerPort,
		Handler:      workerR,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 2)
	go func() {
		log.Info().Str("port", cfg.Port).Str("version", "v1.0.0").Msg("Pushpaka API starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go func() {
		log.Info().Str("port", cfg.WorkerPort).Msg("Pushpaka Worker Management server starting")
		if err := workerSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info().Msg("shutting down API and Worker servers...")
	tunnel.GlobalManager.CloseAll()
	if err := workerSrv.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("worker server shutdown error")
	}
	return srv.Shutdown(shutCtx)
}
