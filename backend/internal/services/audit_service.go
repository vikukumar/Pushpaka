package services

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type AuditService struct {
	repo *repositories.AuditRepository
}

func NewAuditService(repo *repositories.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Log records an audit event asynchronously so it never blocks the request handler.
func (s *AuditService) Log(userID, action, resource, resourceID string, meta map[string]any, ipAddr, userAgent string) {
	go func() {
		metaBytes, _ := json.Marshal(meta)
		entry := &models.AuditLog{
			BaseModel: basemodel.BaseModel{
				ID:        uuid.New().String(),
				CreatedAt: time.Now().UTC(),
			},
			UserID:     userID,
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			Metadata:   string(metaBytes),
			IPAddr:     ipAddr,
			UserAgent:  userAgent,
		}
		_ = s.repo.Create(entry)
	}()
}

func (s *AuditService) List(userID string, limit, offset int) ([]models.AuditLog, error) {
	return s.repo.FindByUserID(userID, limit, offset)
}

func (s *AuditService) ListAll(limit, offset int) ([]models.AuditLog, error) {
	return s.repo.FindAll(limit, offset)
}
