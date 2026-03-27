package repositories

import (
	"errors"

	"gorm.io/gorm"

	"github.com/vikukumar/pushpaka/pkg/basemodel"
	"github.com/vikukumar/pushpaka/pkg/models"
)

var ErrConfigNotFound = errors.New("system config not found")

type SystemConfigRepository struct {
	db *gorm.DB
}

func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

// Get finds a configuration value by key. Returns empty string if not found.
func (r *SystemConfigRepository) Get(key string) (string, error) {
	cfg, err := basemodel.First[models.SystemConfig](r.db, "id = ?", key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrConfigNotFound
		}
		return "", err
	}
	return cfg.Value, nil
}

// Set creates or updates a configuration value by key.
func (r *SystemConfigRepository) Set(key, value string) error {
	cfg := models.SystemConfig{
		ID:    key,
		Value: value,
	}
	return basemodel.Modify(r.db, &cfg)
}
