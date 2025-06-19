package models

import (
	"github.com/zayyadi/finance-tracker/internal/types"
	"gorm.io/gorm"
)

type SavingsGoal struct {
	gorm.Model
	GoalName      string           `json:"goal_name" binding:"required"`
	GoalAmount    float64          `json:"goal_amount" binding:"required"`
	CurrentAmount float64          `json:"current_amount"`
	StartDate     types.CustomDate `json:"start_date"` // Optional, defaults to creation date if not provided
	TargetDate    types.CustomDate `json:"target_date"`
	Notes         string           `json:"notes"`
	// UserID        uint    `json:"user_id"`
}

// SavingsGoalUpdateRequest defines the structure for updating a savings goal.
type SavingsGoalUpdateRequest struct {
	GoalName      *string           `json:"goal_name"`
	GoalAmount    *float64          `json:"goal_amount"`
	CurrentAmount *float64          `json:"current_amount"`
	StartDate     *types.CustomDate `json:"start_date"`
	TargetDate    *types.CustomDate `json:"target_date"`
	Notes         *string           `json:"notes"`
}
