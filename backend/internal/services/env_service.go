package services

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/vikukumar/Pushpaka/internal/models"
	"github.com/vikukumar/Pushpaka/internal/repositories"
)

type EnvService struct {
	envRepo     *repositories.EnvVarRepository
	projectRepo *repositories.ProjectRepository
}

func NewEnvService(
	envRepo *repositories.EnvVarRepository,
	projectRepo *repositories.ProjectRepository,
) *EnvService {
	return &EnvService{
		envRepo:     envRepo,
		projectRepo: projectRepo,
	}
}

func (s *EnvService) Set(userID string, req *models.SetEnvVarRequest) (*models.EnvVarResponse, error) {
	if _, err := s.projectRepo.FindByID(req.ProjectID, userID); err != nil {
		return nil, ErrProjectNotFound
	}

	now := models.NowUTC()
	e := &models.EnvVar{
		ID:        uuid.New().String(),
		ProjectID: req.ProjectID,
		UserID:    userID,
		Key:       req.Key,
		Value:     req.Value,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.envRepo.Upsert(e); err != nil {
		return nil, fmt.Errorf("setting env var: %w", err)
	}

	return &models.EnvVarResponse{
		ID:        e.ID,
		ProjectID: e.ProjectID,
		Key:       e.Key,
		HasValue:  true,
		CreatedAt: e.CreatedAt,
	}, nil
}

func (s *EnvService) List(projectID, userID string) ([]models.EnvVarResponse, error) {
	if _, err := s.projectRepo.FindByID(projectID, userID); err != nil {
		return nil, ErrProjectNotFound
	}

	vars, err := s.envRepo.FindByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	result := make([]models.EnvVarResponse, len(vars))
	for i, v := range vars {
		result[i] = models.EnvVarResponse{
			ID:        v.ID,
			ProjectID: v.ProjectID,
			Key:       v.Key,
			HasValue:  v.Value != "",
			CreatedAt: v.CreatedAt,
		}
	}
	return result, nil
}

func (s *EnvService) Delete(projectID, key, userID string) error {
	if _, err := s.projectRepo.FindByID(projectID, userID); err != nil {
		return ErrProjectNotFound
	}
	return s.envRepo.Delete(projectID, key, userID)
}
