package basemodel

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Connect initializes the database connection based on driver and DSN
func Connect(driver, dsn string, appMode string) (*gorm.DB, error) {
	var dialect gorm.Dialector

	switch driver {
	case "postgres":
		dialect = postgres.Open(dsn)
	case "mysql":
		dialect = mysql.Open(dsn)
	case "sqlite":
		dialect = sqlite.Open(dsn)
	case "sqlserver", "mssql":
		dialect = sqlserver.Open(dsn)
	case "mongodb":
		return nil, fmt.Errorf("MongoDB requires a different driver connection, currently abstracted out of relational basemodel")
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	config := &gorm.Config{
		Logger: NewCustomLogger(appMode),
	}

	db, err := gorm.Open(dialect, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %w", driver, err)
	}

	DB = db
	log.Printf("Successfully connected to %s database", driver)
	return db, nil
}
