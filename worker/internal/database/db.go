package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// New opens a database connection for the given driver.
// driver: "postgres" (default, empty string) or "sqlite".
// For "sqlite" the embedded schema is automatically applied on first open.
func New(driver, dsn string) (*sqlx.DB, error) {
	switch driver {
	case "sqlite":
		return NewSQLite(dsn)
	case "postgres", "":
		return NewPostgres(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %q (choose postgres or sqlite)", driver)
	}
}
