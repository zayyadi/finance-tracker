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
				strings.Contains(storeErr.Error(), "idx_type_period") { // Ensure this index name is correct
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
		SummaryType:     summaryType,
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
	result := s.DB.Where("summary_type = ? AND period_start_date = ?", summaryType, periodStartDate).First(&summary)
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

func (s *SummaryService) calculateTotalForPeriodGORM(userID uint, startDate, endDate time.Time, modelInstance interface{}) (float64, error) {
	var total sql.NullFloat64
	result := s.DB.Model(modelInstance).
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total)

	if result.Error != nil {
		log.Printf("Error calculating total between %s and %s for model %T: %v", startDate, endDate, modelInstance, result.Error)
		return 0, result.Error
	}
	return total.Float64, nil
}

func CalculatePeriodDates(targetDate time.Time, summaryType string) (time.Time, time.Time, error) {
	loc := targetDate.Location()
	var startDate, endDate time.Time

	switch summaryType {
	case "weekly":
		weekday := targetDate.Weekday()
		daysToMonday := int(time.Monday - weekday)
		if weekday == time.Sunday {
			daysToMonday = -6
		}
		startDate = targetDate.AddDate(0, 0, daysToMonday)
		endDate = startDate.AddDate(0, 0, 6)
	case "monthly":
		startDate = time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, -1)
	case "yearly":
		startDate = time.Date(targetDate.Year(), time.January, 1, 0, 0, 0, 0, loc)
		endDate = time.Date(targetDate.Year(), time.December, 31, 0, 0, 0, 0, loc)
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid summary type: %s", summaryType)
	}
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, loc)

	return startDate, endDate, nil
}

// InvalidateSummariesForDate deletes summary records for specified period types based on the itemDate.
func (s *SummaryService) InvalidateSummariesForDate(itemDate time.Time, summaryPeriodTypes []string) error {
	if s.DB == nil {
		return fmt.Errorf("database connection not initialized in SummaryService for invalidation")
	}

	var firstError error
	for _, periodType := range summaryPeriodTypes {
		periodStartDate, _, err := CalculatePeriodDates(itemDate, periodType)
		if err != nil {
			log.Printf("Error calculating period dates for invalidation (type: %s, itemDate: %s): %v", periodType, itemDate.Format("2006-01-02"), err)
			if firstError == nil { // Store the first error encountered
				firstError = fmt.Errorf("failed to calculate period for %s: %w", periodType, err)
			}
			continue // Try to invalidate other periods even if one fails
		}

		result := s.DB.Where("summary_type = ? AND period_start_date = ?", periodType, periodStartDate).Delete(&models.FinancialSummary{})
		if result.Error != nil {
			// Log error, but don't necessarily stop.
			log.Printf("Error deleting summary for invalidation (type: %s, period_start_date: %s): %v", periodType, periodStartDate.Format("2006-01-02"), result.Error)
			if firstError == nil {
				firstError = fmt.Errorf("failed to delete summary for %s (period starting %s): %w", periodType, periodStartDate.Format("2006-01-02"), result.Error)
			}
		} else if result.RowsAffected > 0 {
			log.Printf("Invalidated summary: type %s, period_start_date %s (triggered by item on %s)",
				periodType, periodStartDate.Format("2006-01-02"), itemDate.Format("2006-01-02"))
		}
	}
	return firstError // Return the first error encountered, or nil if all successful
}
