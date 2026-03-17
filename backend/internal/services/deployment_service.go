package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/vikukumar/Pushpaka/internal/models"
	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/queue"
)

const deployJobQueue = "pushpaka:deploy:queue"

var ErrDeploymentNotFound = errors.New("deployment not found")
var ErrQueueUnavailable = errors.New("deployment queue unavailable")

type DeploymentService struct {
	deploymentRepo *repositories.DeploymentRepository
	projectRepo    *repositories.ProjectRepository
	envRepo        *repositories.EnvVarRepository
	domainRepo     *repositories.DomainRepository
	rdb            *redis.Client
	inQueue        *queue.InProcess // non-nil in dev mode (no Redis)
	baseURL        string
}

func NewDeploymentService(
	deploymentRepo *repositories.DeploymentRepository,
	projectRepo *repositories.ProjectRepository,
	envRepo *repositories.EnvVarRepository,
	domainRepo *repositories.DomainRepository,
	rdb *redis.Client,
	inQueue *queue.InProcess,
	baseURL string,
) *DeploymentService {
	svc := &DeploymentService{
		deploymentRepo: deploymentRepo,
		projectRepo:    projectRepo,
		envRepo:        envRepo,
		domainRepo:     domainRepo,
		rdb:            rdb,
		inQueue:        inQueue,
		baseURL:        baseURL,
	}
	// The in-process queue is ephemeral: jobs do not survive process restarts.
	// Any deployment left in "queued" state from a previous run has no
	// corresponding job in the current queue, so fail them immediately.
	// When Redis IS configured the external worker handles its own queue.
	if rdb == nil {
		_ = deploymentRepo.FailStaleQueued(
			"Deployment cancelled: process was restarted before this job was processed. Trigger a new deployment.",
		)
	}
	return svc
}

func (s *DeploymentService) Trigger(userID string, req *models.DeployRequest) (*models.Deployment, error) {
	project, err := s.projectRepo.FindByID(req.ProjectID, userID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	branch := req.Branch
	if branch == "" {
		branch = project.Branch
	}

	now := models.NowUTC()
	imageTag := fmt.Sprintf("pushpaka/%s:%s", project.ID[:8], uuid.New().String()[:8])

	// Determine the public URL for this deployment:
	//   - If the project has a verified custom domain -> https://<domain>
	//   - Otherwise -> <baseURL>/app/<projectID>
	deployURL := s.baseURL + "/app/" + project.ID
	if domains, err := s.domainRepo.FindByProjectID(project.ID); err == nil {
		for _, d := range domains {
			if d.Verified {
				scheme := "http"
				if d.SSLEnabled {
					scheme = "https"
				}
				deployURL = scheme + "://" + d.Domain
				break
			}
		}
	}

	d := &models.Deployment{
		ID:        uuid.New().String(),
		ProjectID: project.ID,
		UserID:    userID,
		CommitSHA: req.CommitSHA,
		Branch:    branch,
		Status:    models.DeploymentQueued,
		ImageTag:  imageTag,
		URL:       deployURL,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.deploymentRepo.Create(d); err != nil {
		return nil, fmt.Errorf("creating deployment record: %w", err)
	}

	// If no queue is available at all, mark the deployment failed immediately
	// so the UI shows a clear error instead of spinning forever.
	if s.rdb == nil && s.inQueue == nil {
		failedAt := models.NowUTC()
		d.Status = models.DeploymentFailed
		d.ErrorMsg = "Deployment worker unavailable: start Pushpaka with -dev for " +
			"the embedded worker, or configure REDIS_URL for a production worker."
		d.FinishedAt = &failedAt
		d.UpdatedAt = failedAt
		_ = s.deploymentRepo.Update(d)
		return d, nil
	}

	// Load env vars
	envVars, err := s.envRepo.FindMapByProjectID(project.ID)
	if err != nil {
		envVars = map[string]string{}
	}

	job := &models.DeploymentJob{
		DeploymentID:    d.ID,
		ProjectID:       project.ID,
		UserID:          userID,
		RepoURL:         project.RepoURL,
		Branch:          branch,
		CommitSHA:       req.CommitSHA,
		BuildCommand:    project.BuildCommand,
		StartCommand:    project.StartCommand,
		Port:            project.Port,
		EnvVars:         envVars,
		ImageTag:        imageTag,
		GitToken:        project.GitToken,
		CPULimit:        project.CPULimit,
		MemoryLimit:     project.MemoryLimit,
		RestartPolicy:   project.RestartPolicy,
		NotificationURL: s.baseURL + "/api/v1/internal/notify",
	}

	payload, err := json.Marshal(job)
	if err != nil {
		return nil, fmt.Errorf("marshaling job: %w", err)
	}

	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.rdb.LPush(ctx, deployJobQueue, payload).Err(); err != nil {
			return nil, fmt.Errorf("queuing job: %w", err)
		}
	} else {
		// In-process queue: faster than Redis, no network round-trip.
		if err := s.inQueue.Push(payload); err != nil {
			return nil, fmt.Errorf("in-process queue: %w", err)
		}
	}

	return d, nil
}

func (s *DeploymentService) List(userID string, limit, offset int) ([]models.Deployment, error) {
	return s.deploymentRepo.FindByUserID(userID, limit, offset)
}

func (s *DeploymentService) ListByProject(projectID, userID string, limit, offset int) ([]models.Deployment, error) {
	// Verify project ownership
	if _, err := s.projectRepo.FindByID(projectID, userID); err != nil {
		return nil, ErrProjectNotFound
	}
	return s.deploymentRepo.FindByProjectID(projectID, limit, offset)
}

func (s *DeploymentService) Get(id string) (*models.Deployment, error) {
	d, err := s.deploymentRepo.FindByID(id)
	if err != nil {
		return nil, ErrDeploymentNotFound
	}
	return d, nil
}

func (s *DeploymentService) Rollback(deploymentID, userID string) (*models.Deployment, error) {
	prev, err := s.deploymentRepo.FindByID(deploymentID)
	if err != nil {
		return nil, ErrDeploymentNotFound
	}

	req := &models.DeployRequest{
		ProjectID: prev.ProjectID,
		Branch:    prev.Branch,
		CommitSHA: prev.CommitSHA,
	}
	return s.Trigger(userID, req)
}
