package basemodel

import "errors"

// AutoMigrate registers models with the DB, checking if tables exist and creating/syncing schemas based on struct tags.
func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return errors.New("database not initialized")
	}
	return DB.AutoMigrate(models...)
}
