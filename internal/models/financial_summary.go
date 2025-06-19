package models

import (
	"time"

	"gorm.io/gorm"
)

// FinancialSummary struct corresponds to the FinancialSummaries table schema.
type FinancialSummary struct {
	gorm.Model
	// UserID          uint      `json:"user_id" gorm:"not null;index;uniqueIndex:idx_user_type_period"` // Removed
	SummaryType     string    `json:"summary_type" gorm:"not null;uniqueIndex:idx_type_period"` // 'weekly', 'monthly', 'yearly'
	PeriodStartDate time.Time `json:"period_start_date" gorm:"not null;uniqueIndex:idx_type_period"`
	PeriodEndDate   time.Time `json:"period_end_date" gorm:"not null"`
	TotalIncome     float64   `json:"total_income" gorm:"not null;default:0"`
	TotalExpenses   float64   `json:"total_expenses" gorm:"not null;default:0"`
	NetBalance      float64   `json:"net_balance" gorm:"not null;default:0"`
}

// SummaryRequest is used for handlers to parse query parameters for summary generation.
// This is not a DB model.
type SummaryRequest struct {
	// For monthly: "YYYY-MM", For yearly: "YYYY", For weekly: "YYYY-MM-DD" (any date within the week)
	// Made optional; if not provided, service logic should default to current period.
	Date string `form:"date"`
}
