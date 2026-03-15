package repositories

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/pushpaka/internal/models"
)

type EnvVarRepository struct {
	db *sqlx.DB
}

func NewEnvVarRepository(db *sqlx.DB) *EnvVarRepository {
	return &EnvVarRepository{db: db}
}

func (r *EnvVarRepository) Upsert(e *models.EnvVar) error {
	e.UpdatedAt = time.Now().UTC()
	query := `
			INSERT INTO environment_variables (id, project_id, user_id, key, value, created_at, updated_at)
			VALUES (:id, :project_id, :user_id, :key, :value, :created_at, :updated_at)
			ON CONFLICT (project_id, key) DO UPDATE
			SET value = excluded.value, updated_at = excluded.updated_at`
	_, err := r.db.NamedExec(query, e)
	return err
}

func (r *EnvVarRepository) FindByProjectID(projectID string) ([]models.EnvVar, error) {
	var envVars []models.EnvVar
	err := r.db.Select(&envVars,
		r.db.Rebind(`SELECT * FROM environment_variables WHERE project_id = ? ORDER BY key ASC`),
		projectID)
	return envVars, err
}

func (r *EnvVarRepository) Delete(projectID, key, userID string) error {
	_, err := r.db.Exec(
		r.db.Rebind(`DELETE FROM environment_variables WHERE project_id = ? AND key = ? AND user_id = ?`),
		projectID, key, userID)
	return err
}

func (r *EnvVarRepository) FindMapByProjectID(projectID string) (map[string]string, error) {
	vars, err := r.FindByProjectID(projectID)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(vars))
	for _, v := range vars {
		m[v.Key] = v.Value
	}
	return m, nil
}
