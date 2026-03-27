package repositories

import (
	"time"

	"gorm.io/gorm"

	"github.com/vikukumar/pushpaka/pkg/basemodel"
	"github.com/vikukumar/pushpaka/pkg/models"
)

type DomainRepository struct {
	db *gorm.DB
}

func NewDomainRepository(db *gorm.DB) *DomainRepository {
	return &DomainRepository{db: db}
}

func (r *DomainRepository) Create(d *models.Domain) error {
	return basemodel.Add(r.db, d)
}

func (r *DomainRepository) FindByProjectID(projectID string) ([]models.Domain, error) {
	return basemodel.Query[models.Domain](r.db, "project_id = ?", projectID)
}

func (r *DomainRepository) FindByUserID(userID string) ([]models.Domain, error) {
	return basemodel.Query[models.Domain](r.db, "user_id = ?", userID)
}

func (r *DomainRepository) FindByDomain(domain string) (*models.Domain, error) {
	return basemodel.First[models.Domain](r.db, "domain = ?", domain)
}

func (r *DomainRepository) SetVerified(id string, verified bool) error {
	now := time.Now().UTC()
	return basemodel.Update[models.Domain](r.db, id, map[string]interface{}{
		"verified":   verified,
		"updated_at": now,
	})
}

func (r *DomainRepository) Delete(id, userID string) error {
	basemodel.EnsureSynced[models.Domain](r.db)
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Domain{}).Error
}
