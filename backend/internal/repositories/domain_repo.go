package repositories

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

type DomainRepository struct {
	db *sqlx.DB
}

func NewDomainRepository(db *sqlx.DB) *DomainRepository {
	return &DomainRepository{db: db}
}

func (r *DomainRepository) Create(d *models.Domain) error {
	query := `
		INSERT INTO domains (id, project_id, user_id, domain, verified, ssl_enabled, created_at, updated_at)
		VALUES (:id, :project_id, :user_id, :domain, :verified, :ssl_enabled, :created_at, :updated_at)`
	_, err := r.db.NamedExec(query, d)
	return err
}

func (r *DomainRepository) FindByProjectID(projectID string) ([]models.Domain, error) {
	var domains []models.Domain
	err := r.db.Select(&domains,
		r.db.Rebind(`SELECT * FROM domains WHERE project_id = ? ORDER BY created_at DESC`), projectID)
	return domains, err
}

func (r *DomainRepository) FindByUserID(userID string) ([]models.Domain, error) {
	var domains []models.Domain
	err := r.db.Select(&domains,
		r.db.Rebind(`SELECT * FROM domains WHERE user_id = ? ORDER BY created_at DESC`), userID)
	return domains, err
}

func (r *DomainRepository) FindByDomain(domain string) (*models.Domain, error) {
	var d models.Domain
	err := r.db.Get(&d, r.db.Rebind(`SELECT * FROM domains WHERE domain = ?`), domain)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DomainRepository) SetVerified(id string, verified bool) error {
	_, err := r.db.Exec(
		r.db.Rebind(`UPDATE domains SET verified = ?, updated_at = ? WHERE id = ?`),
		verified, time.Now().UTC(), id)
	return err
}

func (r *DomainRepository) Delete(id, userID string) error {
	_, err := r.db.Exec(r.db.Rebind(`DELETE FROM domains WHERE id = ? AND user_id = ?`), id, userID)
	return err
}
