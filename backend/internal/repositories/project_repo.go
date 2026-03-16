package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

type ProjectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(p *models.Project) error {
	query := `
		INSERT INTO projects (id, user_id, name, repo_url, branch, build_command, start_command, port, framework, status, is_private, git_token, created_at, updated_at)
		VALUES (:id, :user_id, :name, :repo_url, :branch, :build_command, :start_command, :port, :framework, :status, :is_private, :git_token, :created_at, :updated_at)`
	_, err := r.db.NamedExec(query, p)
	return err
}

func (r *ProjectRepository) FindByUserID(userID string) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Select(&projects, r.db.Rebind(`SELECT * FROM projects WHERE user_id = ? ORDER BY created_at DESC`), userID)
	return projects, err
}

func (r *ProjectRepository) FindByID(id, userID string) (*models.Project, error) {
	var p models.Project
	err := r.db.Get(&p, r.db.Rebind(`SELECT * FROM projects WHERE id = ? AND user_id = ?`), id, userID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) Update(p *models.Project) error {
	query := `
		UPDATE projects
		SET name = :name, repo_url = :repo_url, branch = :branch,
		    build_command = :build_command, start_command = :start_command,
		    port = :port, framework = :framework, status = :status,
		    is_private = :is_private, git_token = :git_token, updated_at = :updated_at
		WHERE id = :id AND user_id = :user_id`
	_, err := r.db.NamedExec(query, p)
	return err
}

// FindByIDInternal returns the project including the git_token (for worker jobs).
// Only use this server-side — never marshal the result directly to an API response.
func (r *ProjectRepository) FindByIDInternal(id string) (*models.Project, error) {
	var p models.Project
	err := r.db.Get(&p, r.db.Rebind(`SELECT * FROM projects WHERE id = ?`), id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) Delete(id, userID string) error {
	_, err := r.db.Exec(r.db.Rebind(`DELETE FROM projects WHERE id = ? AND user_id = ?`), id, userID)
	return err
}
