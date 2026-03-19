package repositories

import (
	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return basemodel.Add(r.db, user)
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	return basemodel.First[models.User](r.db, "email = ?", email)
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	return basemodel.Get[models.User](r.db, id)
}

func (r *UserRepository) FindByAPIKey(apiKey string) (*models.User, error) {
	return basemodel.First[models.User](r.db, "api_key = ?", apiKey)
}

func (r *UserRepository) Update(user *models.User) error {
	return basemodel.Modify(r.db, user)
}
