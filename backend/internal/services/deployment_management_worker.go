package services

import (
	"context"
	"log"
	"time"

	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type DeploymentManagementWorker struct {
	projectRepo  *repositories.ProjectRepository
	deployRepo   *repositories.DeploymentRepository
	dmRepo       *repositories.DeploymentManagementRepository
	dmService    *DeploymentManagementService
	tickInterval time.Duration
	stopChan     chan bool
	isStopped    bool
}

func NewDeploymentManagementWorker(
	projectRepo *repositories.ProjectRepository,
	deployRepo *repositories.DeploymentRepository,
	dmRepo *repositories.DeploymentManagementRepository,
	dmService *DeploymentManagementService,
	tickInterval time.Duration,
) *DeploymentManagementWorker {
	return &DeploymentManagementWorker{
		projectRepo:  projectRepo,
		deployRepo:   deployRepo,
		dmRepo:       dmRepo,
		dmService:    dmService,
		tickInterval: tickInterval,
		stopChan:     make(chan bool),
		isStopped:    true,
	}
}

// ============= Worker Lifecycle =============

// Start begins the deployment management worker
func (w *DeploymentManagementWorker) Start() {
	if !w.isStopped {
		log.Println("DeploymentManagementWorker is already running")
		return
	}

	w.isStopped = false
	go w.run()
	log.Println("DeploymentManagementWorker started")
}

// Stop halts the deployment management worker
func (w *DeploymentManagementWorker) Stop() {
	if w.isStopped {
		log.Println("DeploymentManagementWorker is already stopped")
		return
	}

	w.stopChan <- true
	w.isStopped = true
	log.Println("DeploymentManagementWorker stopped")
}

// run is the main worker loop
func (w *DeploymentManagementWorker) run() {
	ticker := time.NewTicker(w.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			w.processDeploymentTasks(ctx)
			cancel()

		case <-w.stopChan:
			return
		}
	}
}

// ============= Deployment Tasks =============

// processDeploymentTasks executes all pending deployment management tasks
func (w *DeploymentManagementWorker) processDeploymentTasks(ctx context.Context) {
	log.Println("[DeploymentWorker] Processing deployment tasks...")

	// Process pending actions
	w.processPendingActions(ctx)

	// Check deployment health
	w.checkDeploymentHealth(ctx)

	// Cleanup old backups
	w.cleanupOldBackups(ctx)

	// Process pending project clones
	w.processPendingProjectClones(ctx)

	log.Println("[DeploymentWorker] Deployment tasks completed")
}

// processPendingActions processes all pending deployment actions
func (w *DeploymentManagementWorker) processPendingActions(ctx context.Context) {
	// TODO: Get all pending actions from database
	// TODO: Execute action based on type (start/stop/restart/sync/etc)
	// TODO: Update action status

	log.Println("[DeploymentWorker] Processing pending actions...")
}

// checkDeploymentHealth checks health of all running deployments
func (w *DeploymentManagementWorker) checkDeploymentHealth(ctx context.Context) {
	log.Println("[DeploymentWorker] Checking deployment health...")

	// TODO: Get all running deployment instances
	// TODO: Make health check request to each
	// TODO: Update health_status in database
	// TODO: If unhealthy and restart_count < max, trigger restart action
	// TODO: If healthy, reset restart_count to 0

	// Example:
	// for each instance {
	//   if instance.Status == "running" {
	//     health := checkHealth(instance.Port)
	//     if health.Status != "healthy" && instance.RestartCount < MAX_RESTARTS {
	//       recordAction(instance, "restart")
	//     }
	//   }
	// }
}

// cleanupOldBackups removes backups exceeding the project limit
func (w *DeploymentManagementWorker) cleanupOldBackups(ctx context.Context) {
	log.Println("[DeploymentWorker] Cleaning up old backups...")

	// TODO: Get all projects
	// TODO: For each project, get backup count
	// TODO: If backup count > MaxBackups, delete oldest backups
	// TODO: Log deleted backups

	// Example:
	// projects := getAllProjects()
	// for each project {
	//   backups := getBackupsByProject(project.ID)
	//   if len(backups) > project.MaxBackups {
	//     oldestBackups := getOldestBackups(project.ID, len(backups) - project.MaxBackups)
	//     for each oldestBackup {
	//       deleteBackupDirectory(oldestBackup)
	//       deleteBackupRecord(oldestBackup.ID)
	//     }
	//   }
	// }
}

// processPendingProjectClones clones repositories for new projects
func (w *DeploymentManagementWorker) processPendingProjectClones(ctx context.Context) {
	log.Println("[DeploymentWorker] Processing pending project clones...")

	// TODO: Get all projects with GitClonePath == ""
	// TODO: For each project, initialize project clone
	// TODO: Update project with clone path
	// TODO: Log cloned projects

	// Example:
	// projects := getProjectsWithoutClones()
	// for each project {
	//   if project.GitClonePath == "" {
	//     clonePath, err := dmService.InitializeProjectClone(ctx, project)
	//     if err != nil {
	//       log.Printf("Failed to clone for project %s: %v", project.ID, err)
	//     } else {
	//       log.Printf("Cloned project %s to %s", project.ID, clonePath)
	//     }
	//   }
	// }
}

// ============= Event Handlers (Can be called directly for immediate processing) =============

// OnProjectCreated handles project creation event
func (w *DeploymentManagementWorker) OnProjectCreated(ctx context.Context, project *models.Project) {
	log.Printf("[DeploymentWorker] Project created event: %s", project.ID)

	// Clone repository immediately
	if project.GitClonePath == "" {
		clonePath, err := w.dmService.InitializeProjectClone(ctx, project)
		if err != nil {
			log.Printf("[DeploymentWorker] Failed to clone project %s: %v", project.ID, err)
		} else {
			log.Printf("[DeploymentWorker] Project %s cloned to %s", project.ID, clonePath)
		}
	}
}

// OnProjectUpdated handles project update event (repo/branch change)
func (w *DeploymentManagementWorker) OnProjectUpdated(ctx context.Context, project *models.Project, oldBranch string) {
	log.Printf("[DeploymentWorker] Project updated event: %s", project.ID)

	// Refresh clone if branch changed
	if project.Branch != oldBranch {
		if err := w.dmService.RefreshProjectClone(ctx, project, oldBranch); err != nil {
			log.Printf("[DeploymentWorker] Failed to refresh clone for project %s: %v", project.ID, err)
		} else {
			log.Printf("[DeploymentWorker] Project %s clone refreshed", project.ID)
		}
	}
}

// OnDeploymentTriggered handles deployment trigger event
func (w *DeploymentManagementWorker) OnDeploymentTriggered(ctx context.Context, deployment *models.Deployment) {
	log.Printf("[DeploymentWorker] Deployment triggered event: %s", deployment.ID)

	// TODO: Execute full deployment workflow
	// 1. Get project
	// 2. Capture code signature
	// 3. Create deployment instances (main/testing)
	// 4. Backup current main if exists
	// 5. Start deployment
	// 6. Health check
	// 7. Update status
}

// OnActionTriggered handles action trigger event
func (w *DeploymentManagementWorker) OnActionTriggered(ctx context.Context, action *models.DeploymentAction) {
	log.Printf("[DeploymentWorker] Action triggered: %s for deployment %s", action.Action, action.DeploymentID)

	// TODO: Execute action based on type
	// switch action.Action {
	// case models.DeploymentActionStart:
	//   startDeployment(ctx, action)
	// case models.DeploymentActionStop:
	//   stopDeployment(ctx, action)
	// case models.DeploymentActionRestart:
	//   restartDeployment(ctx, action)
	// case models.DeploymentActionRetry:
	//   retryDeployment(ctx, action)
	// case models.DeploymentActionRollback:
	//   rollbackDeployment(ctx, action)
	// case models.DeploymentActionSync:
	//   syncDeployment(ctx, action)
	// }
}

// ============= Utility Methods =============

// HealthCheck performs a health check on a single instance
func (w *DeploymentManagementWorker) healthCheck(port int) (string, error) {
	// TODO: Make HTTP request to localhost:port/health
	// TODO: Parse response and return status
	// TODO: Handle timeout and connection errors

	// Example:
	// timeout := 5 * time.Second
	// client := &http.Client{Timeout: timeout}
	// resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", port))
	// if err != nil {
	//   return "unhealthy", err
	// }
	// defer resp.Body.Close()
	// if resp.StatusCode == 200 {
	//   return "healthy", nil
	// }
	// return "unhealthy", nil

	return "unknown", nil
}

// RestartDeploymentIfUnhealthy restarts deployment if health check fails
func (w *DeploymentManagementWorker) RestartDeploymentIfUnhealthy(ctx context.Context, instance *models.DeploymentInstance, maxRestarts int) error {
	if instance.Status != "running" {
		return nil
	}

	status, err := w.healthCheck(instance.Port)
	if err != nil {
		log.Printf("[DeploymentWorker] Health check error for instance %s: %v", instance.ID, err)
		status = "unhealthy"
	}

	if status != "healthy" {
		if instance.RestartCount >= maxRestarts {
			log.Printf("[DeploymentWorker] Instance %s exceeded max restarts (%d)", instance.ID, maxRestarts)
			return nil
		}

		log.Printf("[DeploymentWorker] Restarting unhealthy instance %s (restart #%d)", instance.ID, instance.RestartCount+1)

		// TODO: Trigger restart action
		// action, err := w.dmService.RecordAction(
		//   instance.DeploymentID,
		//   instance.ID,
		//   instance.ProjectID,
		//   "system",
		//   models.DeploymentActionRestart,
		// )
		// if err != nil {
		//   return err
		// }

		// Update restart count
		instance.RestartCount++
		instance.HealthStatus = "attempting_restart"
		instance.UpdatedAt = time.Now().UTC()

		// TODO: Update instance in database
		// return w.dmRepo.UpdateDeploymentInstance(instance)
	} else {
		// Reset restart count if healthy
		instance.RestartCount = 0
		instance.HealthStatus = "healthy"
		instance.UpdatedAt = time.Now().UTC()

		// TODO: Update instance in database
	}

	return nil
}
