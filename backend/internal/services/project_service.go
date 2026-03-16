package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/vikukumar/Pushpaka/internal/models"
	"github.com/vikukumar/Pushpaka/internal/repositories"
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

	now := models.NowUTC()
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
		IsPrivate:    req.IsPrivate,
		GitToken:     req.GitToken,
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

func (s *ProjectService) Update(id, userID string, req *models.UpdateProjectRequest) (*models.Project, error) {
	p, err := s.projectRepo.FindByID(id, userID)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Branch != "" {
		p.Branch = req.Branch
	}
	if req.BuildCommand != "" {
		p.BuildCommand = req.BuildCommand
	}
	if req.StartCommand != "" {
		p.StartCommand = req.StartCommand
	}
	if req.Port > 0 {
		p.Port = req.Port
	}
	if req.Framework != "" {
		p.Framework = req.Framework
	}
	p.IsPrivate = req.IsPrivate
	// Only update the token when a new one is explicitly provided.
	if req.GitToken != "" {
		p.GitToken = req.GitToken
	}
	p.UpdatedAt = models.NowUTC()

	if err := s.projectRepo.Update(p); err != nil {
		return nil, fmt.Errorf("updating project: %w", err)
	}
	return p, nil
}

func (s *ProjectService) Delete(id, userID string) error {
	return s.projectRepo.Delete(id, userID)
}
