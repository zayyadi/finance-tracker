package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zayyadi/finance-tracker/internal/database"
	"github.com/zayyadi/finance-tracker/internal/models"

	// "github.com/lib/pq" // No longer needed for pq.Error with GORM generally
	"gorm.io/gorm"
)

// SummaryService provides methods for generating and retrieving financial summaries using GORM.
type SummaryService struct {
	DB *gorm.DB
}

// NewSummaryService creates a new SummaryService with a GORM database connection.
func NewSummaryService(db *gorm.DB) *SummaryService {
	if db == nil {
		log.Println("Warning: NewSummaryService called with nil DB, attempting to use global GetDB()")
		db = database.GetDB()
	}
	return &SummaryService{DB: db}
}

// GetOrCreateFinancialSummary fetches an existing summary or generates a new one.
func (s *SummaryService) GetOrCreateFinancialSummary(summaryType string, targetDate time.Time) (*models.FinancialSummary, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in SummaryService")
	}

	periodStartDate, periodEndDate, err := CalculatePeriodDates(targetDate, summaryType)
	if err != nil {
		return nil, fmt.Errorf("erhttps://docs.google.com/forms/d/e/1FAIpQLSfAdZXGGMT716G9Gfz7884A8ywZPAH2NRoENLwEpiJJoHBo4Q/viewform?s=35ror calculating period dates: %w", err)
	}

	summary, err := s.fetchSummaryFromDB(summaryType, periodStartDate) // Removed userID
	if err == nil && summary != nil {
		return summary, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error fetching existing summary, type %s, date %s: %v", summaryType, periodStartDate.Format("2006-01-02"), err)
		return nil, fmt.Errorf("error retrieving existing summary: %w", err)
	}

	totalIncome, err := s.calculateTotalForPeriodGORM(0, periodStartDate, periodEndDate, &models.Income{}) // Removed userID
	if err != nil {
		return nil, fmt.Errorf("error calculating total income: %w", err)
	}

	totalExpenses, err := s.calculateTotalForPeriodGORM(0, periodStartDate, periodEndDate, &models.Expense{}) // Removed userID
	if err != nil {
		return nil, fmt.Errorf("error calculating total expenses: %w", err)
	}

	netBalance := totalIncome - totalExpenses

	newSummary := &models.FinancialSummary{
		// UserID:          userID, // Removed
		SummaryType:     summaryType,
		PeriodStartDate: periodStartDate,
		PeriodEndDate:   periodEndDate,
		TotalIncome:     totalIncome,
		TotalExpenses:   totalExpenses,
		NetBalance:      netBalance,
	}

	storedSummary, err := s.storeSummaryInDB(newSummary)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") &&
			strings.Contains(err.Error(), "idx_type_period") { // Updated unique index name
			log.Printf("Unique constraint violation for summary type %s, date %s. Attempting to re-fetch.", summaryType, periodStartDate.Format("2006-01-02"))
			return s.fetchSummaryFromDB(summaryType, periodStartDate) // Removed userID
		}
		log.Printf("Error storing new summary type %s, date %s: %v", summaryType, periodStartDate.Format("2006-01-02"), err)
		return nil, fmt.Errorf("error storing new summary: %w", err)
	}
	return storedSummary, nil
}

func (s *SummaryService) fetchSummaryFromDB(summaryType string, periodStartDate time.Time) (*models.FinancialSummary, error) { // Removed userID
	var summary models.FinancialSummary
	result := s.DB.Where("summary_type = ? AND period_start_date = ?", summaryType, periodStartDate).First(&summary) // Removed userID
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &summary, nil
}

func (s *SummaryService) storeSummaryInDB(summary *models.FinancialSummary) (*models.FinancialSummary, error) {
	result := s.DB.Create(summary)
	if result.Error != nil {
		return nil, result.Error
	}
	return summary, nil
}

// calculateTotalForPeriodGORM calculates sum of 'amount' for a given table (model) within a date range.
// modelInstance should be a pointer to an empty struct of the model type (e.g., &models.Income{}).
func (s *SummaryService) calculateTotalForPeriodGORM(userID uint, startDate, endDate time.Time, modelInstance interface{}) (float64, error) {
	var total sql.NullFloat64 // Use sql.NullFloat64 to handle potential NULL sum from DB
	// Removed Where("user_id = ?", userID) as the application is now single-user
	result := s.DB.Model(modelInstance).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)"). // Ensure 0 if no records
		Scan(&total)

	if result.Error != nil {
		// It's better to log the specific model being queried if possible, but modelInstance is interface{}
		log.Printf("Error calculating total for user %d between %s and %s: %v", userID, startDate, endDate, result.Error)
		return 0, result.Error
	}
	return total.Float64, nil // Return 0.0 if sum was NULL (due to COALESCE)
}

// CalculatePeriodDates calculates the start and end dates for a given summary type and target date.
// It's exported for testing purposes and potentially for use elsewhere if needed.
func CalculatePeriodDates(targetDate time.Time, summaryType string) (time.Time, time.Time, error) {
	loc := targetDate.Location() // Use the location of the targetDate for calculations
	var startDate, endDate time.Time

	switch summaryType {
	case "weekly":
		// Assuming week starts on Monday and ends on Sunday
		weekday := targetDate.Weekday()
		// Adjust to Monday (Sunday is 0, Monday is 1, ..., Saturday is 6)
		daysToMonday := int(time.Monday - weekday)
		if weekday == time.Sunday { // In Go, Sunday is 0, so if we want Monday as start, Sunday needs special handling
			daysToMonday = -6
		}
		startDate = targetDate.AddDate(0, 0, daysToMonday)
		endDate = startDate.AddDate(0, 0, 6) // Sunday
	case "monthly":
		startDate = time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, -1) // Last day of the month
	case "yearly":
		startDate = time.Date(targetDate.Year(), time.January, 1, 0, 0, 0, 0, loc)
		endDate = time.Date(targetDate.Year(), time.December, 31, 0, 0, 0, 0, loc)
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid summary type: %s", summaryType)
	}
	// Normalize to midnight for date-only comparisons if necessary, though DB DATE type handles this.
	// For consistency, let's ensure they are at the beginning of the day in their respective timezone.
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, loc)

	return startDate, endDate, nil
}
