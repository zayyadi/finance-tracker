package models

import (
	"github.com/zayyadi/finance-tracker/internal/types"
	"gorm.io/gorm"
)

// Expense struct corresponds to the Expenses table schema.
type Expense struct {
	gorm.Model
	// UserID   uint    `json:"user_id" gorm:"not null;index"` // Removed
	Amount   float64          `json:"amount" binding:"required,gt=0" gorm:"not null;default:0"`
	Category string           `json:"category" binding:"required" gorm:"not null"`
	Date     types.CustomDate `json:"date" binding:"required" gorm:"not null"`
	Note     string           `json:"note,omitempty"`
}

// ExpenseCreateRequest defines the expected request body for creating an expense.
type ExpenseCreateRequest struct {
	Amount   float64          `json:"amount" binding:"required,gt=0"`
	Category string           `json:"category" binding:"required"`
	Date     types.CustomDate `json:"date" binding:"required"`
	Note     string           `json:"note,omitempty"`
}

// ExpenseUpdateRequest defines the expected request body for updating an expense.
// All fields are optional, allowing partial updates.
type ExpenseUpdateRequest struct {
	Amount   *float64          `json:"amount,omitempty" binding:"omitempty,gt=0"`
	Category *string           `json:"category,omitempty"`
	Date     *types.CustomDate `json:"date,omitempty"`
	Note     *string           `json:"note,omitempty"`
}
