package database

import (
	"fmt"
	"log"
	"os"

	// "database/sql" // No longer directly using sql.DB for GetDB

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// _ "github.com/lib/pq" // GORM's driver will handle this
)

var gormDB *gorm.DB

// ConnectDB initializes the GORM database connection using environment variables.
func ConnectDB() error {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")       // It's good practice to make port configurable
	dbSSLMode := os.Getenv("DB_SSLMODE") // Moved up

	// Validate essential environment variables
	if dbHost == "" {
		return fmt.Errorf("DB_HOST environment variable is not set or is empty")
	}
	// DB_USER is typically required. The error log indicates it was set ('zayyadev') in this instance.
	if dbUser == "" {
		return fmt.Errorf("DB_USER environment variable is not set or is empty")
	}
	if dbName == "" {
		return fmt.Errorf("DB_NAME environment variable is not set or is empty")
	}

	if dbPort == "" {
		dbPort = "5432" // Default PostgreSQL port
	}

	if dbSSLMode == "" {
		dbSSLMode = "disable" // Default to disable for local dev, require for prod
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode)

	var err error
	gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("error opening database with GORM: %w", err)
	}

	// Optional: Ping the database to ensure connection is live
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("error getting underlying sql.DB from GORM: %w", err)
	}
	err = sqlDB.Ping()
	if err != nil {
		// sqlDB.Close() // GORM handles closing the underlying connection
		return fmt.Errorf("error connecting to database (ping failed): %w", err)
	}

	log.Println("Successfully connected to the database using GORM!")
	return nil
}

// GetDB returns the active GORM database connection.
// It's the caller's responsibility to ensure ConnectDB has been called successfully.
func GetDB() *gorm.DB {
	if gormDB == nil {
		log.Println("Warning: GetDB (GORM) called before database connection was initialized or connection failed.")
		// Attempt to connect if not already connected.
		// This is a fallback, ideally ConnectDB is called at application startup.
		if err := ConnectDB(); err != nil {
			log.Printf("Failed to connect to database (GORM) in GetDB: %v", err)
			return nil // Return nil if connection fails
		}
	}
	return gormDB
}

// CloseDB closes the GORM database connection.
// It should be called when the application is shutting down.
func CloseDB() {
	if gormDB != nil {
		sqlDB, err := gormDB.DB()
		if err != nil {
			log.Printf("Error getting underlying sql.DB from GORM for closing: %v", err)
			return
		}
		err = sqlDB.Close()
		if err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("Database connection closed (GORM).")
		}
	}
}
