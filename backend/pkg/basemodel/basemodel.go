package basemodel

import (
	"time"

	"gorm.io/gorm"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

// BaseModel provides standard fields and soft-delete capabilities for all models
type BaseModel struct {
	ID        string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Get retrieves a single record by its primary key ID
func Get[T any](db *gorm.DB, id string) (*T, error) {
	var dest T
	err := db.First(&dest, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &dest, nil
}

// Add inserts a new record
func Add[T any](db *gorm.DB, record *T) error {
	return db.Create(record).Error
}

// Modify updates an existing record
func Modify[T any](db *gorm.DB, record *T) error {
	return db.Save(record).Error
}

// Update explicit fields using a map
func Update[T any](db *gorm.DB, id string, data map[string]interface{}) error {
	var model T
	return db.Model(&model).Where("id = ?", id).Updates(data).Error
}

// Delete soft-deletes a record
func Delete[T any](db *gorm.DB, id string) error {
	var dest T
	return db.Where("id = ?", id).Delete(&dest).Error
}

// HardDelete permanently removes a record
func HardDelete[T any](db *gorm.DB, id string) error {
	var dest T
	return db.Unscoped().Where("id = ?", id).Delete(&dest).Error
}

// Query allows executing custom queries with conditions
func Query[T any](db *gorm.DB, query string, args ...interface{}) ([]T, error) {
	var dest []T
	err := db.Where(query, args...).Find(&dest).Error
	return dest, err
}

// First query wrapper to return single record
func First[T any](db *gorm.DB, query string, args ...interface{}) (*T, error) {
	var dest T
	err := db.Where(query, args...).First(&dest).Error
	if err != nil {
		return nil, err
	}
	return &dest, nil
}

// Execute allows raw SQL execution for more complex operations
func Execute(db *gorm.DB, sql string, args ...interface{}) error {
	return db.Exec(sql, args...).Error
}
