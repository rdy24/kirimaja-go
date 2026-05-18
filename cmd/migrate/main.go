// Command migrate runs database schema migration and optional seeding.
//
//	go run ./cmd/migrate            # migrate + seed
//	go run ./cmd/migrate -seed=false # migrate only
//	go run ./cmd/migrate -only=seed  # seed only
package main

import (
	"flag"
	"log"

	"kirimaja-go/internal/config"
	"kirimaja-go/internal/database"
)

func main() {
	seed := flag.Bool("seed", true, "run seeder after migration")
	only := flag.String("only", "", "run only one step: migrate | seed")
	flag.Parse()

	cfg := config.Load()
	db, err := database.Connect(cfg.DatabaseURL, cfg.IsProduction())
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	switch *only {
	case "migrate":
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	case "seed":
		if err := database.Seed(db); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
	default:
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		if *seed {
			if err := database.Seed(db); err != nil {
				log.Fatalf("Seeding failed: %v", err)
			}
		}
	}

	log.Println("Done")
}
