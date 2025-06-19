package database

import (
	"fmt"
	"log"

	"github.com/zayyadi/finance-tracker/internal/models" // Assuming this is the correct path to your models
	"gorm.io/gorm"
)

// AutoMigrateModels automatically migrates GORM models to the database schema.
func AutoMigrateModels(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("cannot auto-migrate with a nil DB instance")
	}
	log.Println("Starting GORM auto-migration...")

	// Add all your GORM models here
	err := db.AutoMigrate(
		// &models.User{}, // User model removed
		&models.Income{},
		&models.Expense{},
		&models.Savings{},
		&models.Debt{},
		&models.FinancialSummary{},
	)

	if err != nil {
		log.Printf("Error during GORM auto-migration: %v", err)
		return fmt.Errorf("gorm auto-migration failed: %w", err)
	}

	log.Println("GORM auto-migration completed successfully.")
	return nil
}
