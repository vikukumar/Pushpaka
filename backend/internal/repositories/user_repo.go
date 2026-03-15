package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, email, name, password_hash, api_key, role, created_at, updated_at)
		VALUES (:id, :email, :name, :password_hash, :api_key, :role, :created_at, :updated_at)`
	_, err := r.db.NamedExec(query, user)
	return err
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user, r.db.Rebind(`SELECT * FROM users WHERE email = ?`), email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user, r.db.Rebind(`SELECT * FROM users WHERE id = ?`), id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByAPIKey(apiKey string) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user, r.db.Rebind(`SELECT * FROM users WHERE api_key = ?`), apiKey)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users SET name = :name, updated_at = :updated_at WHERE id = :id`
	_, err := r.db.NamedExec(query, user)
	return err
}
