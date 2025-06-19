package models

import (
	"github.com/zayyadi/finance-tracker/internal/types"
	"gorm.io/gorm"
)

// Debt represents a debt owed by or to the user.
type Debt struct {
	gorm.Model
	// UserID      uint      `json:"user_id" gorm:"not null;index"` // Removed
	DebtorName  string           `json:"debtor_name" binding:"required" gorm:"not null"`
	Description string           `json:"description,omitempty"`
	Amount      float64          `json:"amount" binding:"required,gt=0" gorm:"not null;default:0"`
	DueDate     types.CustomDate `json:"due_date" binding:"required" gorm:"not null"`
	Status      string           `json:"status" binding:"required,oneof=Pending Paid Overdue" gorm:"not null;default:'Pending'"`
}

// DebtCreateRequest is used for creating a new debt record.
type DebtCreateRequest struct {
	DebtorName  string           `json:"debtor_name" binding:"required"`
	Description *string          `json:"description,omitempty"`
	Amount      float64          `json:"amount" binding:"required,gt=0"`
	DueDate     types.CustomDate `json:"due_date" binding:"required"`
	Status      *string          `json:"status,omitempty" binding:"omitempty,oneof=Pending Paid Overdue"` // Defaults to 'Pending' in service
}

// DebtUpdateRequest is used for updating an existing debt record.
type DebtUpdateRequest struct {
	DebtorName  *string           `json:"debtor_name,omitempty"`
	Description *string           `json:"description,omitempty"` // Pointer to allow explicitly setting to empty vs. not providing
	Amount      *float64          `json:"amount,omitempty" binding:"omitempty,gt=0"`
	DueDate     *types.CustomDate `json:"due_date,omitempty"`
	Status      *string           `json:"status,omitempty" binding:"omitempty,oneof=Pending Paid Overdue"`
}
