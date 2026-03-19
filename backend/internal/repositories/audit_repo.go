package repositories

import (
	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(log *models.AuditLog) error {
	return basemodel.Add(r.db, log)
}

func (r *AuditRepository) FindByUserID(userID string, limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, err
}

func (r *AuditRepository) FindAll(limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, err
}
