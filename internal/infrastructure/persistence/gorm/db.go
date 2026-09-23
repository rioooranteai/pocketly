package persistence

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pocketly/internal/infrastructure/config"
)

/*
models lists every GORM model that has a table in the schema. Migrate
and CheckSchema both read from it, so a new model only needs to be
registered here.
*/
var models = []any{&UserModel{}, &TransactionModel{}, &TransactionItemModel{}}

/*
Connect opens a SQLite database connection using the path from Config.
It does not touch the schema — run Migrate (via cmd/migrate) for that,
so schema changes are applied deliberately rather than on every
server start.
*/
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return db, nil
}

/*
Migrate runs AutoMigrate for every registered model, creating missing
tables and columns. It never drops columns or data.
*/
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

/*
CheckSchema reports an error if any registered model's table is
missing. The API server calls it at startup so an unmigrated database
fails fast with a clear message instead of returning 500s on the
first request.
*/
func CheckSchema(db *gorm.DB) error {
	for _, model := range models {
		if !db.Migrator().HasTable(model) {
			return fmt.Errorf("database schema is missing tables, run: go run ./cmd/migrate")
		}
	}

	return nil
}

/*
Close closes the underlying database connection pool.
*/
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
