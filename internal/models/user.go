package models

import "gorm.io/gorm"

// User represents a user in the system.
// Even if not directly used by all services in a single-user app context,
// it's needed for schema completeness if other tables (like income/expense
// in the main DB schema) have foreign keys to it.
// For testing analytics_service, GORM's AutoMigrate will create this table.
type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Email        string `gorm:"type:varchar(100);uniqueIndex;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	// Add any other fields that might be part of your User model
}
