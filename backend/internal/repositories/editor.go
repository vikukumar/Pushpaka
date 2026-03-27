package repositories

import (
	"github.com/vikukumar/pushpaka/pkg/models"
	"gorm.io/gorm"
)

type EditorStateRepository struct {
	db *gorm.DB
}

func NewEditorStateRepository(db *gorm.DB) *EditorStateRepository {
	return &EditorStateRepository{db: db}
}

func (r *EditorStateRepository) Get(userID, projectID string) (*models.UserEditorState, error) {
	var state models.UserEditorState
	err := r.db.Where("user_id = ? AND project_id = ?", userID, projectID).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *EditorStateRepository) Save(state *models.UserEditorState) error {
	var existing models.UserEditorState
	err := r.db.Where("user_id = ? AND project_id = ?", state.UserID, state.ProjectID).First(&existing).Error
	if err != nil {
		// Create new
		return r.db.Create(state).Error
	}
	// Update existing
	state.ID = existing.ID
	return r.db.Save(state).Error
}
