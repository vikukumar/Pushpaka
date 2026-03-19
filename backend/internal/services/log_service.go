package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
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
		BaseModel: basemodel.BaseModel{
			ID:        uuid.New().String(),
			CreatedAt: time.Now().UTC(),
		},
		DeploymentID: deploymentID,
		Level:        level,
		Stream:       stream,
		Message:      message,
	}
	if err := s.logRepo.Create(l); err != nil {
		return fmt.Errorf("appending log: %w", err)
	}
	return nil
}
