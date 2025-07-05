// File: cmd/seed/main.go
package main

import (
	"al/connection"
	"al/models"
	"al/seeders"
	"al/utils"
	"flag"
	"log"
	"os"
	"time"
)

func main() {
	// Command line flags
	var (
		clean = flag.Bool("clean", false, "Clean all seeder data")
		seed  = flag.Bool("seed", false, "Run all seeders")
	)
	flag.Parse()

	// Initialize environment
	time.LoadLocation(os.Getenv("APP_TIMEZONE"))
	utils.ValidationTranslationInit()

	// Initialize database connection
	connection.InitDB()

	// Auto migrate to ensure tables exist
	connection.DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Otp{},
		&models.Setting{},
	)

	// Initialize seeder
	seeder := seeders.NewSeeder(connection.DB)

	switch {
	case *clean:
		log.Println("Cleaning seeder data...")
		if err := seeder.CleanAll(); err != nil {
			log.Fatal("Failed to clean seeder data:", err)
		}
		
	case *seed:
		log.Println("Running seeders...")
		if err := seeder.RunAll(); err != nil {
			log.Fatal("Failed to run seeders:", err)
		}
		
	default:
		log.Println("Usage:")
		log.Println("  go run cmd/seed/main.go --seed    # Run all seeders")
		log.Println("  go run cmd/seed/main.go --clean   # Clean all seeder data")
	}
}