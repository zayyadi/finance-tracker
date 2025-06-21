package models

import (
	"github.com/zayyadi/finance-tracker/internal/database"
	"gorm.io/gorm"
	"github.com/zayyadi/finance-tracker/internal/database" // Added for CustomDate
)

// Savings represents a savings goal.
type Savings struct {
	gorm.Model
	// UserID        uint       `json:"user_id" gorm:"not null;index"` // Removed
	GoalName      string                `json:"goal_name" binding:"required" gorm:"not null"`
	GoalAmount    float64               `json:"goal_amount" binding:"required,gt=0" gorm:"not null;default:0"`
	CurrentAmount float64               `json:"current_amount" binding:"gte=0" gorm:"not null;default:0"`
	StartDate     *database.CustomDate `json:"start_date,omitempty" gorm:"default:null;type:date"`
	TargetDate    *database.CustomDate `json:"target_date,omitempty" gorm:"default:null;type:date"`
	Notes         string                `json:"notes,omitempty"`

}

// SavingsCreateRequest is used for creating a new savings goal.
type SavingsCreateRequest struct {
	GoalName      string                `json:"goal_name" binding:"required"`
	GoalAmount    float64               `json:"goal_amount" binding:"required,gt=0"`
	CurrentAmount *float64              `json:"current_amount,omitempty" binding:"omitempty,gte=0"` // Optional, defaults to 0 in service
	StartDate     *database.CustomDate `json:"start_date,omitempty"`
	TargetDate    *database.CustomDate `json:"target_date,omitempty"`
	Notes         *string               `json:"notes,omitempty"`
}

// SavingsUpdateRequest is used for updating an existing savings goal.
// All fields are optional.
type SavingsUpdateRequest struct {
	GoalName      *string               `json:"goal_name,omitempty"`
	GoalAmount    *float64              `json:"goal_amount,omitempty" binding:"omitempty,gt=0"`
	CurrentAmount *float64              `json:"current_amount,omitempty" binding:"omitempty,gte=0"`
	StartDate     *database.CustomDate `json:"start_date,omitempty"` // Use pointer to distinguish between not provided and explicit null
	TargetDate    *database.CustomDate `json:"target_date,omitempty"` // Use pointer
	Notes         *string               `json:"notes,omitempty"`     // Use pointer

}
