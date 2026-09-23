package main

import (
	"log"

	"pocketly/internal/infrastructure/config"
	persistence "pocketly/internal/infrastructure/persistence/gorm"
)

/*
main applies the database schema and exits. It is kept separate from
the API server so schema changes run as an explicit deploy step,
before the new server version starts, rather than implicitly on every
server boot.
*/
func main() {
	cfg := config.Load()

	db, err := persistence.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := persistence.Close(db); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	if err := persistence.Migrate(db); err != nil {
		log.Fatal(err)
	}

	log.Printf("migration complete: %s", cfg.DBPath)
}
