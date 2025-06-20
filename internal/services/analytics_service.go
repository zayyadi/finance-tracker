package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/zayyadi/finance-tracker/internal/models"
	"gorm.io/gorm"
)

// AnalyticsService provides methods for generating financial analytics.
type AnalyticsService struct {
	DB *gorm.DB
}

// NewAnalyticsService creates a new AnalyticsService with a GORM database connection.
func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{DB: db}
}

// GetExpenseBreakdownByCategory calculates expense breakdown by category for a given month.
func (s *AnalyticsService) GetExpenseBreakdownByCategory(targetDate time.Time) ([]models.CategoryExpenseStat, error) {
	if s.DB == nil {
		return nil, errors.New("database connection not initialized in AnalyticsService")
	}
	startDate := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
	// Calculate endDate as the first day of the next month, then subtract one day to get the last day of the target month.
	// This avoids issues with varying month lengths.
	endDate := startDate.AddDate(0, 1, 0).AddDate(0, 0, -1)


	var stats []models.CategoryExpenseStat
	result := s.DB.Model(&models.Expense{}).
		Select("category, SUM(amount) as total_amount").
		Where("date BETWEEN ? AND ?", startDate, endDate). // Assuming 'date' column exists in Expense model
		Group("category").
		Order("total_amount DESC").
		Scan(&stats)
	if result.Error != nil {
		log.Printf("Error getting expense breakdown by category for %s: %v", startDate.Format("2006-01"), result.Error)
		return nil, result.Error
	}
	return stats, nil
}

// GetIncomeExpenseTrend calculates income and expense trends for the last numMonths.
func (s *AnalyticsService) GetIncomeExpenseTrend(numMonths int) ([]models.MonthlyTrendStat, error) {
	if s.DB == nil {
		return nil, errors.New("database connection not initialized in AnalyticsService")
	}
	var trend []models.MonthlyTrendStat
	today := time.Now()

	for i := 0; i < numMonths; i++ {
		targetMonthDate := today.AddDate(0, -i, 0)
		monthStartDate := time.Date(targetMonthDate.Year(), targetMonthDate.Month(), 1, 0, 0, 0, 0, targetMonthDate.Location())
		// Calculate monthEndDate similar to GetExpenseBreakdownByCategory for robustness
		monthEndDate := monthStartDate.AddDate(0, 1, 0).AddDate(0,0,-1)


		totalIncome, err := s.calculateTotalForPeriod(monthStartDate, monthEndDate, &models.Income{})
		if err != nil {
			return nil, fmt.Errorf("error calculating income for %s: %w", monthStartDate.Format("2006-01"), err)
		}

		totalExpenses, err := s.calculateTotalForPeriod(monthStartDate, monthEndDate, &models.Expense{})
		if err != nil {
			return nil, fmt.Errorf("error calculating expenses for %s: %w", monthStartDate.Format("2006-01"), err)
		}

		trend = append(trend, models.MonthlyTrendStat{
			Month:         monthStartDate.Format("2006-01"), // Format as YYYY-MM
			TotalIncome:   totalIncome,
			TotalExpenses: totalExpenses,
		})
	}

	// Reverse the trend so it's in chronological order (oldest to newest)
	for j, k := 0, len(trend)-1; j < k; j, k = j+1, k-1 {
		trend[j], trend[k] = trend[k], trend[j]
	}
	return trend, nil
}

// calculateTotalForPeriod is a helper function to calculate sum of 'amount' for a given model.
func (s *AnalyticsService) calculateTotalForPeriod(startDate, endDate time.Time, modelInstance interface{}) (float64, error) {
	var total sql.NullFloat64 // Use sql.NullFloat64 for cases where sum might be null.
	result := s.DB.Model(modelInstance).
		Where("date BETWEEN ? AND ?", startDate, endDate). // Ensure 'date' column exists in the model
		Select("COALESCE(SUM(amount), 0)").               // Return 0 if no records or sum is NULL
		Scan(&total)

	if result.Error != nil {
		// Log the error with model type for better debugging
		log.Printf("Error calculating total between %s and %s for model %T: %v", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), modelInstance, result.Error)
		return 0, result.Error
	}
	return total.Float64, nil // Returns 0.0 if sum was NULL (handled by COALESCE)
}
