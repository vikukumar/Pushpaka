package repositories

import (
	"time"

	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
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
	var domains []models.Domain
	err := r.db.Where("project_id = ?", projectID).Order("created_at desc").Find(&domains).Error
	return domains, err
}

func (r *DomainRepository) FindByUserID(userID string) ([]models.Domain, error) {
	var domains []models.Domain
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&domains).Error
	return domains, err
}

func (r *DomainRepository) FindByDomain(domain string) (*models.Domain, error) {
	return basemodel.First[models.Domain](r.db, "domain = ?", domain)
}

func (r *DomainRepository) SetVerified(id string, verified bool) error {
	return r.db.Model(&models.Domain{}).Where("id = ?", id).Updates(map[string]interface{}{
		"verified":   verified,
		"updated_at": time.Now().UTC(),
	}).Error
}

func (r *DomainRepository) Delete(id, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Domain{}).Error
}
