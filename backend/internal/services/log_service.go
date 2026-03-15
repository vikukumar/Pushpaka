package services

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/vikukumar/Pushpaka/internal/models"
	"github.com/vikukumar/Pushpaka/internal/repositories"
)

type LogService struct {
	logRepo *repositories.LogRepository
}

func NewLogService(logRepo *repositories.LogRepository) *LogService {
	return &LogService{logRepo: logRepo}
}

func (s *LogService) GetByDeployment(deploymentID string) ([]models.DeploymentLog, error) {
	return s.logRepo.FindByDeploymentID(deploymentID)
}

func (s *LogService) Append(deploymentID, level, stream, message string) error {
	l := &models.DeploymentLog{
		ID:           uuid.New().String(),
		DeploymentID: deploymentID,
		Level:        level,
		Stream:       stream,
		Message:      message,
		CreatedAt:    models.NowUTC(),
	}
	if err := s.logRepo.Create(l); err != nil {
		return fmt.Errorf("appending log: %w", err)
	}
	return nil
}
