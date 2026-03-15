package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourusername/pushpaka/internal/models"
	"github.com/yourusername/pushpaka/internal/repositories"
)

var ErrProjectNotFound = errors.New("project not found")

type ProjectService struct {
	projectRepo *repositories.ProjectRepository
}

func NewProjectService(projectRepo *repositories.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

func (s *ProjectService) Create(userID string, req *models.CreateProjectRequest) (*models.Project, error) {
	branch := req.Branch
	if branch == "" {
		branch = "main"
	}
	port := req.Port
	if port == 0 {
		port = 3000
	}

	now := time.Now().UTC()
	p := &models.Project{
		ID:           uuid.New().String(),
		UserID:       userID,
		Name:         req.Name,
		RepoURL:      req.RepoURL,
		Branch:       branch,
		BuildCommand: req.BuildCommand,
		StartCommand: req.StartCommand,
		Port:         port,
		Framework:    req.Framework,
		Status:       "inactive",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.projectRepo.Create(p); err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}
	return p, nil
}

func (s *ProjectService) List(userID string) ([]models.Project, error) {
	return s.projectRepo.FindByUserID(userID)
}

func (s *ProjectService) Get(id, userID string) (*models.Project, error) {
	p, err := s.projectRepo.FindByID(id, userID)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	return p, nil
}

func (s *ProjectService) Delete(id, userID string) error {
	return s.projectRepo.Delete(id, userID)
}
