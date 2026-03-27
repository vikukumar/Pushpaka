package repositories

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/vikukumar/pushpaka/pkg/basemodel"
	"github.com/vikukumar/pushpaka/pkg/models"
)

type EnvVarRepository struct {
	db *gorm.DB
}

func NewEnvVarRepository(db *gorm.DB) *EnvVarRepository {
	return &EnvVarRepository{db: db}
}

func (r *EnvVarRepository) Upsert(e *models.EnvVar) error {
	basemodel.EnsureSynced[models.EnvVar](r.db)
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(e).Error
}

func (r *EnvVarRepository) FindByProjectID(projectID string) ([]models.EnvVar, error) {
	return basemodel.Query[models.EnvVar](r.db, "project_id = ?", projectID)
}

func (r *EnvVarRepository) Delete(projectID, key, userID string) error {
	basemodel.EnsureSynced[models.EnvVar](r.db)
	return r.db.Where("project_id = ? AND key = ? AND user_id = ?", projectID, key, userID).Delete(&models.EnvVar{}).Error
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
