package services

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/vikukumar/Pushpaka/internal/models"
	"github.com/vikukumar/Pushpaka/internal/repositories"
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
			ID:         uuid.New().String(),
			UserID:     userID,
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			Metadata:   string(metaBytes),
			IPAddr:     ipAddr,
			UserAgent:  userAgent,
			CreatedAt:  models.Time{Time: time.Now().UTC()},
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
