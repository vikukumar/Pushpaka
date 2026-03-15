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
)

const deployJobQueue = "pushpaka:deploy:queue"

var ErrDeploymentNotFound = errors.New("deployment not found")

type DeploymentService struct {
	deploymentRepo *repositories.DeploymentRepository
	projectRepo    *repositories.ProjectRepository
	envRepo        *repositories.EnvVarRepository
	rdb            *redis.Client
}

func NewDeploymentService(
	deploymentRepo *repositories.DeploymentRepository,
	projectRepo *repositories.ProjectRepository,
	envRepo *repositories.EnvVarRepository,
	rdb *redis.Client,
) *DeploymentService {
	return &DeploymentService{
		deploymentRepo: deploymentRepo,
		projectRepo:    projectRepo,
		envRepo:        envRepo,
		rdb:            rdb,
	}
}

func (s *DeploymentService) Trigger(userID string, req *models.DeployRequest) (*models.Deployment, error) {
	if s.rdb == nil {
		return nil, errors.New("deployment queue unavailable: REDIS_URL not configured")
	}
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

	d := &models.Deployment{
		ID:        uuid.New().String(),
		ProjectID: project.ID,
		UserID:    userID,
		CommitSHA: req.CommitSHA,
		Branch:    branch,
		Status:    models.DeploymentQueued,
		ImageTag:  imageTag,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.deploymentRepo.Create(d); err != nil {
		return nil, fmt.Errorf("creating deployment record: %w", err)
	}

	// Load env vars
	envVars, err := s.envRepo.FindMapByProjectID(project.ID)
	if err != nil {
		envVars = map[string]string{}
	}

	job := &models.DeploymentJob{
		DeploymentID: d.ID,
		ProjectID:    project.ID,
		UserID:       userID,
		RepoURL:      project.RepoURL,
		Branch:       branch,
		CommitSHA:    req.CommitSHA,
		BuildCommand: project.BuildCommand,
		StartCommand: project.StartCommand,
		Port:         project.Port,
		EnvVars:      envVars,
		ImageTag:     imageTag,
	}

	payload, err := json.Marshal(job)
	if err != nil {
		return nil, fmt.Errorf("marshaling job: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.rdb.LPush(ctx, deployJobQueue, payload).Err(); err != nil {
		return nil, fmt.Errorf("queuing job: %w", err)
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
