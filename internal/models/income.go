package models

import (
	"github.com/zayyadi/finance-tracker/internal/database" // Corrected import
	"gorm.io/gorm"
)

// Income struct corresponds to the Income table schema.
type Income struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt
	// UserID     uint   `json:"user_id" gorm:"not null;index"` // Removed for single-user mode
	Amount   float64            `json:"amount" binding:"required,gt=0" gorm:"not null;default:0"`
	Category string             `json:"category" binding:"required" gorm:"not null"`
	Date     database.CustomDate `json:"date" binding:"required" gorm:"not null"`
	Note     string             `json:"note,omitempty"` // Allow empty, GORM handles it
}

// IncomeCreateRequest defines the expected request body for creating income,
// excluding fields that should be set by the server (ID, UserID, CreatedAt, UpdatedAt).
type IncomeCreateRequest struct {
	Amount   float64            `json:"amount" binding:"required,gt=0"`
	Category string             `json:"category" binding:"required"`
	Date     database.CustomDate `json:"date" binding:"required"`
	Note     string             `json:"note,omitempty"` // Keep as string, GORM handles empty string fine
}

// IncomeUpdateRequest defines the expected request body for updating income.
// All fields are optional, allowing partial updates. GORM handles this well with struct updates.
// Using pointers ensures that only provided fields are updated and can distinguish between
// a zero value (e.g. 0 for amount) and a field not being provided.
type IncomeUpdateRequest struct {
	Amount   *float64            `json:"amount,omitempty" binding:"omitempty,gt=0"`
	Category *string             `json:"category,omitempty"`
	Date     *database.CustomDate `json:"date,omitempty"`
	Note     *string             `json:"note,omitempty"`
}
