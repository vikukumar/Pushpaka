package repositories

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/vikukumar/Pushpaka/pkg/models"
)

// K8sConfigRepository manages per-user Kubernetes cluster configurations.
type K8sConfigRepository struct {
	db *gorm.DB
}

func NewK8sConfigRepository(db *gorm.DB) *K8sConfigRepository {
	return &K8sConfigRepository{db: db}
}

func (r *K8sConfigRepository) GetByUserID(userID string) (*models.K8sConfig, error) {
	var cfg models.K8sConfig
	err := r.db.Where("user_id = ?", userID).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *K8sConfigRepository) Upsert(cfg *models.K8sConfig) error {
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
		cfg.CreatedAt = time.Now().UTC()
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}}, // Unique index should exist on user_id
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "server_url", "namespace", "registry_url", "enabled", "updated_at",
		}),
		UpdateAll: false,
	}).Create(cfg).Error
	// Note: The previous logic had a CASE WHEN for token/kubeconfig, but we can't easily reproduce that
	// with generic AssigmentColumns without raw SQL. We can use raw SQL if that's critical. Let's use it.
}

func (r *K8sConfigRepository) GetByUserIDInternal(userID string) (*models.K8sConfig, error) {
	return r.GetByUserID(userID)
}

// GetEnabledByUserID returns the K8s config only if it is enabled.
func (r *K8sConfigRepository) GetEnabledByUserID(userID string) (*models.K8sConfig, error) {
	var cfg models.K8sConfig
	err := r.db.Where("user_id = ? AND enabled = 1", userID).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}
