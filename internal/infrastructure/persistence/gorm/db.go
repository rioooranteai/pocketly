package persistence

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pocketly/internal/infrastructure/config"
)

/*
Connect opens a SQLite database connection using the path from Config,
then runs AutoMigrate to ensure all registered models have matching
tables. It returns a ready-to-use *gorm.DB, or an error if either the
connection or migration step fails.
*/
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	err = db.AutoMigrate(&UserModel{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
