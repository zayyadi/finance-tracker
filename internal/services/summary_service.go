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

// GetOrCreateFinancialSummary fetches an existing summary or generates a new one based on viewType.
// viewType can be "overall", "income", "expenses". Other types are not yet implemented.
// Overall summaries are fetched from/stored in DB. View-specific summaries are calculated on the fly.
func (s *SummaryService) GetOrCreateFinancialSummary(summaryType string, targetDate time.Time, viewType string) (*models.FinancialSummary, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in SummaryService")
	}

	if viewType == "" {
		viewType = "overall" // Default to overall
	}

	periodStartDate, periodEndDate, err := CalculatePeriodDates(targetDate, summaryType)
	if err != nil {
		// Corrected the error message from the original code that had a URL in it.
		return nil, fmt.Errorf("error calculating period dates: %w", err)
	}

	// Handle "overall" view - fetch from DB or calculate and store
	if viewType == "overall" {
		summary, err := s.fetchSummaryFromDB(summaryType, periodStartDate)
		if err == nil && summary != nil {
			return summary, nil // Found existing overall summary
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Error fetching existing summary (overall), type %s, date %s: %v", summaryType, periodStartDate.Format("2006-01-02"), err)
			return nil, fmt.Errorf("error retrieving existing overall summary: %w", err)
		}

		// Existing overall summary not found, calculate it
		totalIncome, errIncome := s.calculateTotalForPeriodGORM(0, periodStartDate, periodEndDate, &models.Income{})
		if errIncome != nil {
			return nil, fmt.Errorf("error calculating total income for overall summary: %w", errIncome)
		}
		totalExpenses, errExpenses := s.calculateTotalForPeriodGORM(0, periodStartDate, periodEndDate, &models.Expense{})
		if errExpenses != nil {
			return nil, fmt.Errorf("error calculating total expenses for overall summary: %w", errExpenses)
		}
		netBalance := totalIncome - totalExpenses

		newSummary := &models.FinancialSummary{
			SummaryType:     summaryType,
			PeriodStartDate: periodStartDate,
			PeriodEndDate:   periodEndDate,
			TotalIncome:     totalIncome,
			TotalExpenses:   totalExpenses,
			NetBalance:      netBalance,
		}
		// Attempt to store the new overall summary
		storedSummary, storeErr := s.storeSummaryInDB(newSummary)
		if storeErr != nil {
			// Handle potential race condition where another request created the summary in the meantime
			if strings.Contains(storeErr.Error(), "duplicate key value violates unique constraint") &&
				strings.Contains(storeErr.Error(), "idx_type_period") {
				log.Printf("Unique constraint violation for overall summary type %s, date %s during store. Re-fetching.", summaryType, periodStartDate.Format("2006-01-02"))
				return s.fetchSummaryFromDB(summaryType, periodStartDate)
			}
			log.Printf("Error storing new overall summary type %s, date %s: %v", summaryType, periodStartDate.Format("2006-01-02"), storeErr)
			return nil, fmt.Errorf("error storing new overall summary: %w", storeErr)
		}
		return storedSummary, nil
	}

	// Handle view-specific calculations (not stored in DB)
	var totalIncome float64
	var totalExpenses float64
	var calcErr error

	if viewType == "income" {
		totalIncome, calcErr = s.calculateTotalForPeriodGORM(0, periodStartDate, periodEndDate, &models.Income{})
		totalExpenses = 0 // Expenses are zero for income-only view
	} else if viewType == "expenses" {
		totalIncome = 0 // Income is zero for expenses-only view
		totalExpenses, calcErr = s.calculateTotalForPeriodGORM(0, periodStartDate, periodEndDate, &models.Expense{})
	} else if viewType == "savings" || viewType == "debts" {
		// Placeholder for future implementation
		return nil, fmt.Errorf("viewType '%s' not yet implemented", viewType)
	} else {
		return nil, fmt.Errorf("invalid viewType '%s'", viewType)
	}

	if calcErr != nil {
		return nil, fmt.Errorf("error calculating totals for view '%s': %w", viewType, calcErr)
	}

	netBalance := totalIncome - totalExpenses
	// Create a non-persistent FinancialSummary object for the specific view
	viewSummary := &models.FinancialSummary{
		SummaryType:     summaryType, // Could also add viewType to SummaryType string if needed for frontend
		PeriodStartDate: periodStartDate,
		PeriodEndDate:   periodEndDate,
		TotalIncome:     totalIncome,
		TotalExpenses:   totalExpenses,
		NetBalance:      netBalance,
	}
	return viewSummary, nil
}

func (s *SummaryService) fetchSummaryFromDB(summaryType string, periodStartDate time.Time) (*models.FinancialSummary, error) {
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
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)"). // Ensure 0 if no records
		Scan(&total)

	if result.Error != nil {
		// It's better to log the specific model being queried if possible, but modelInstance is interface{}
		log.Printf("Error calculating total between %s and %s for model %T: %v", startDate, endDate, modelInstance, result.Error)
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
