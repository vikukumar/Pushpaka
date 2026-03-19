package basemodel

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var (
	DB           *gorm.DB
	DBReady      atomic.Bool
	syncedModels sync.Map
)

// IsDBReady turns the atomic flag into a boolean for handlers.
func IsDBReady() bool {
	return DBReady.Load()
}

// SetDBReady updates the atomic readiness flag.
func SetDBReady(ready bool) {
	DBReady.Store(ready)
}

// EnsureSynced checks if a model's table has been migration-checked in this session.
// If not, it runs a targeted AutoMigrate. Thread-safe.
func EnsureSynced[T any](db *gorm.DB) {
	// If global migration (background) is done, we don't need to check individually.
	if IsDBReady() {
		return
	}

	var model T
	typeName := fmt.Sprintf("%T", model)

	if _, loaded := syncedModels.Load(typeName); !loaded {
		// First-time access: sync the table.
		// Targeted AutoMigrate is fast as it only checks one structure.
		if err := db.AutoMigrate(&model); err != nil {
			log.Printf("[DB] Target sync failed for %s: %v", typeName, err)
			return
		}
		syncedModels.Store(typeName, true)
	}
}

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
