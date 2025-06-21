package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zayyadi/finance-tracker/internal/database"
	"github.com/zayyadi/finance-tracker/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupSummaryTestDB initializes an in-memory SQLite database for summary service testing.
func setupSummaryTestDB(t *testing.T) *gorm.DB {
	// Use a unique DSN for each test to ensure isolation with cache=shared
	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", t.Name(), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err, "Failed to connect to in-memory SQLite")

	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })

	// Drop tables first for a clean state
	db.Exec("DROP TABLE IF EXISTS FinancialSummary")
	db.Exec("DROP TABLE IF EXISTS Expenses")
	db.Exec("DROP TABLE IF EXISTS Income")
	db.Exec("DROP TABLE IF EXISTS Users")

	// Auto-migrate schemas based on GORM structs.
	err = db.AutoMigrate(&models.User{}, &models.Income{}, &models.Expense{}, &models.FinancialSummary{})
	assert.NoError(t, err, "Failed to auto-migrate models")

	// Optional: Create a dummy user if needed
	// db.Create(&models.User{Username: "testuser", Email: "test@example.com", PasswordHash: "hash"})

	return db
}

// seedDataForSummaryTest populates income and expense data for a given period.
func seedDataForSummaryTest(t *testing.T, db *gorm.DB, dateForIncome time.Time, incomeAmount float64, dateForExpense time.Time, expenseAmount float64) {
	if incomeAmount > 0 {
		income := models.Income{Amount: incomeAmount, Category: "Test Income", Date: database.CustomDate{Time: dateForIncome}}
		err := db.Create(&income).Error
		assert.NoError(t, err, "Failed to seed income: %+v", income)
	}
	if expenseAmount > 0 {
		expense := models.Expense{Amount: expenseAmount, Category: "Test Expense", Date: database.CustomDate{Time: dateForExpense}}
		err := db.Create(&expense).Error
		assert.NoError(t, err, "Failed to seed expense: %+v", expense)
	}
}

// seedFinancialSummary creates and stores a FinancialSummary record for testing.
func seedFinancialSummary(t *testing.T, db *gorm.DB, summaryType string, forDateForPeriodCalc time.Time, totalIncome float64, totalExpenses float64) models.FinancialSummary {
	periodStartDate, periodEndDate, err := CalculatePeriodDates(forDateForPeriodCalc, summaryType)
	assert.NoError(t, err, "Failed to calculate period dates for seeding summary")

	summary := models.FinancialSummary{
		SummaryType:     summaryType,
		PeriodStartDate: periodStartDate,
		PeriodEndDate:   periodEndDate,
		TotalIncome:     totalIncome,
		TotalExpenses:   totalExpenses,
		NetBalance:      totalIncome - totalExpenses,
	}
	errCreate := db.Create(&summary).Error
	assert.NoError(t, errCreate, "Failed to seed financial summary: %+v", summary)
	return summary
}


func TestCalculatePeriodDates(t *testing.T) {
	loc, _ := time.LoadLocation("UTC") // Use a consistent timezone for tests

	testCases := []struct {
		name            string
		targetDate      time.Time
		summaryType     string
		expectedStart   time.Time
		expectedEnd     time.Time
		expectError     bool
		expectedErrorMsg string
	}{
		// Monthly tests
		{
			name:          "Monthly_MidMonth",
			targetDate:    time.Date(2023, time.November, 15, 0, 0, 0, 0, loc),
			summaryType:   "monthly",
			expectedStart: time.Date(2023, time.November, 1, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2023, time.November, 30, 0, 0, 0, 0, loc),
			expectError:   false,
		},
		{
			name:          "Monthly_StartOfMonth",
			targetDate:    time.Date(2023, time.February, 1, 0, 0, 0, 0, loc),
			summaryType:   "monthly",
			expectedStart: time.Date(2023, time.February, 1, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2023, time.February, 28, 0, 0, 0, 0, loc), // Non-leap year
			expectError:   false,
		},
		{
			name:          "Monthly_LeapYear",
			targetDate:    time.Date(2024, time.February, 10, 0, 0, 0, 0, loc),
			summaryType:   "monthly",
			expectedStart: time.Date(2024, time.February, 1, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2024, time.February, 29, 0, 0, 0, 0, loc), // Leap year
			expectError:   false,
		},

		// Weekly tests (assuming Monday is the start of the week)
		{
			name:          "Weekly_MidWeek_Wednesday", // Wednesday
			targetDate:    time.Date(2023, time.November, 15, 0, 0, 0, 0, loc), // 2023-11-15 is a Wednesday
			summaryType:   "weekly",
			expectedStart: time.Date(2023, time.November, 13, 0, 0, 0, 0, loc), // Monday
			expectedEnd:   time.Date(2023, time.November, 19, 0, 0, 0, 0, loc), // Sunday
			expectError:   false,
		},
		{
			name:          "Weekly_Monday",
			targetDate:    time.Date(2023, time.November, 13, 0, 0, 0, 0, loc), // Monday
			summaryType:   "weekly",
			expectedStart: time.Date(2023, time.November, 13, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2023, time.November, 19, 0, 0, 0, 0, loc),
			expectError:   false,
		},
		{
			name:          "Weekly_Sunday",
			targetDate:    time.Date(2023, time.November, 19, 0, 0, 0, 0, loc), // Sunday
			summaryType:   "weekly",
			expectedStart: time.Date(2023, time.November, 13, 0, 0, 0, 0, loc), // Monday of that week
			expectedEnd:   time.Date(2023, time.November, 19, 0, 0, 0, 0, loc), // Sunday
			expectError:   false,
		},
		 {
			name:          "Weekly_AcrossMonthBoundary",
			targetDate:    time.Date(2023, time.October, 30, 0, 0, 0, 0, loc), // Monday Oct 30
			summaryType:   "weekly",
			expectedStart: time.Date(2023, time.October, 30, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2023, time.November, 5, 0, 0, 0, 0, loc),
			expectError:   false,
		},


		// Yearly tests
		{
			name:          "Yearly_AnyDate",
			targetDate:    time.Date(2023, time.July, 15, 0, 0, 0, 0, loc),
			summaryType:   "yearly",
			expectedStart: time.Date(2023, time.January, 1, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2023, time.December, 31, 0, 0, 0, 0, loc),
			expectError:   false,
		},

		// Error cases
		{
			name:            "InvalidSummaryType",
			targetDate:      time.Now(),
			summaryType:     "daily", // Assuming 'daily' is not supported
			expectError:     true,
			expectedErrorMsg: "invalid summary type: daily",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			startDate, endDate, err := CalculatePeriodDates(tc.targetDate, tc.summaryType)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. Start: %v, End: %v", startDate, endDate)
				} else if err.Error() != tc.expectedErrorMsg {
					t.Errorf("Expected error message '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error, but got: %v", err)
				}
				if !startDate.Equal(tc.expectedStart) {
					t.Errorf("Expected start date %v, but got %v", tc.expectedStart, startDate)
				}
				if !endDate.Equal(tc.expectedEnd) {
					t.Errorf("Expected end date %v, but got %v", tc.expectedEnd, endDate)
				}
			}
		})
	}
}

// --- Tests for GetOrCreateFinancialSummary with viewType ---

func TestGetOrCreateFinancialSummary_ViewOverall_Monthly_NoExisting(t *testing.T) {
	db := setupSummaryTestDB(t)
	service := NewSummaryService(db)
	targetDate := time.Date(2023, time.April, 10, 0, 0, 0, 0, time.UTC)
	periodStart := time.Date(2023, time.April, 1, 0, 0, 0, 0, time.UTC)

	seedDataForSummaryTest(t, db, periodStart.AddDate(0,0,5), 1000, periodStart.AddDate(0,0,10), 300)
	seedDataForSummaryTest(t, db, periodStart.AddDate(0,0,15), 500, periodStart.AddDate(0,0,20), 200) // Income: 1500, Expense: 500

	summary, err := service.GetOrCreateFinancialSummary("monthly", targetDate, "overall")
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 1500.0, summary.TotalIncome)
	assert.Equal(t, 500.0, summary.TotalExpenses)
	assert.Equal(t, 1000.0, summary.NetBalance)
	assert.Equal(t, periodStart, summary.PeriodStartDate)

	// Verify it was stored in DB
	var dbSummary models.FinancialSummary
	result := db.Where("summary_type = ? AND period_start_date = ?", "monthly", periodStart).First(&dbSummary)
	assert.NoError(t, result.Error)
	assert.Equal(t, 1500.0, dbSummary.TotalIncome)
}

func TestGetOrCreateFinancialSummary_ViewOverall_Monthly_Existing(t *testing.T) {
	db := setupSummaryTestDB(t)
	service := NewSummaryService(db)
	targetDate := time.Date(2023, time.May, 10, 0, 0, 0, 0, time.UTC)
	periodStart := time.Date(2023, time.May, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2023, time.May, 31, 0,0,0,0,time.UTC)

	// Pre-store a summary
	preStoredSummary := models.FinancialSummary{
		SummaryType:     "monthly",
		PeriodStartDate: periodStart,
		PeriodEndDate:   periodEnd,
		TotalIncome:     2000,
		TotalExpenses:   800,
		NetBalance:      1200,
	}
	db.Create(&preStoredSummary)

	// Seed some data anyway, to ensure the service prefers the stored one for "overall"
	seedDataForSummaryTest(t, db, periodStart.AddDate(0,0,5), 100, periodStart.AddDate(0,0,10), 50)


	summary, err := service.GetOrCreateFinancialSummary("monthly", targetDate, "overall")
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 2000.0, summary.TotalIncome, "Should return pre-stored income")
	assert.Equal(t, 800.0, summary.TotalExpenses, "Should return pre-stored expenses")
	assert.Equal(t, 1200.0, summary.NetBalance, "Should return pre-stored net balance")
}

func TestGetOrCreateFinancialSummary_ViewIncome_Monthly(t *testing.T) {
	db := setupSummaryTestDB(t)
	service := NewSummaryService(db)
	targetDate := time.Date(2023, time.June, 10, 0, 0, 0, 0, time.UTC)
	periodStart := time.Date(2023, time.June, 1, 0, 0, 0, 0, time.UTC)

	seedDataForSummaryTest(t, db, periodStart.AddDate(0,0,5), 1200, periodStart.AddDate(0,0,10), 300)

	summary, err := service.GetOrCreateFinancialSummary("monthly", targetDate, "income")
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 1200.0, summary.TotalIncome)
	assert.Equal(t, 0.0, summary.TotalExpenses)
	assert.Equal(t, 1200.0, summary.NetBalance)

	// Verify it was NOT stored as a new specific record (overall might be there if called before)
	var dbSummaries []models.FinancialSummary
	db.Where("summary_type = ? AND period_start_date = ?", "monthly", periodStart).Find(&dbSummaries)
	// If an overall summary was created by another test or logic, it might exist.
	// The key is that this "income" view didn't create a *second* record or overwrite overall with partial data.
	// For simplicity, we check that if a record exists, its expenses are not 0 (unless actual expenses were 0).
	// This test assumes no prior "overall" record for June.
	assert.Equal(t, 0, len(dbSummaries), "Income-only view should not be stored, and no overall for June existed")
}

func TestGetOrCreateFinancialSummary_ViewExpenses_Monthly(t *testing.T) {
	db := setupSummaryTestDB(t)
	service := NewSummaryService(db)
	targetDate := time.Date(2023, time.July, 10, 0, 0, 0, 0, time.UTC)
	periodStart := time.Date(2023, time.July, 1, 0, 0, 0, 0, time.UTC)

	seedDataForSummaryTest(t, db, periodStart.AddDate(0,0,5), 1500, periodStart.AddDate(0,0,10), 450)

	summary, err := service.GetOrCreateFinancialSummary("monthly", targetDate, "expenses")
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 0.0, summary.TotalIncome)
	assert.Equal(t, 450.0, summary.TotalExpenses)
	assert.Equal(t, -450.0, summary.NetBalance)

	var dbSummaries []models.FinancialSummary
	db.Where("summary_type = ? AND period_start_date = ?", "monthly", periodStart).Find(&dbSummaries)
	assert.Equal(t, 0, len(dbSummaries), "Expense-only view should not be stored, and no overall for July existed")
}

// Similar tests should be added for "weekly" and "yearly" periodTypes if full coverage is desired.
// For brevity in this example, only monthly is extensively tested for view types.
// Adding one weekly example:
func TestGetOrCreateFinancialSummary_ViewIncome_Weekly(t *testing.T) {
	db := setupSummaryTestDB(t)
	service := NewSummaryService(db)
	// Test for a week: Monday 2023-Oct-02 to Sunday 2023-Oct-08
	targetDateInWeek := time.Date(2023, time.October, 4, 0, 0, 0, 0, time.UTC) // A Wednesday

	seedDataForSummaryTest(t, db, time.Date(2023, time.October, 3, 0,0,0,0,time.UTC), 200, time.Date(2023, time.October, 5,0,0,0,0,time.UTC), 50) // In week
	seedDataForSummaryTest(t, db, time.Date(2023, time.September, 30,0,0,0,0,time.UTC), 1000, time.Date(2023, time.October, 10,0,0,0,0,time.UTC), 500) // Outside week

	summary, err := service.GetOrCreateFinancialSummary("weekly", targetDateInWeek, "income")
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 200.0, summary.TotalIncome)
	assert.Equal(t, 0.0, summary.TotalExpenses)
	assert.Equal(t, 200.0, summary.NetBalance)
}


func TestGetOrCreateFinancialSummary_ViewNotImplemented(t *testing.T) {
	db := setupSummaryTestDB(t)
	service := NewSummaryService(db)
	targetDate := time.Now()

	_, errSavings := service.GetOrCreateFinancialSummary("monthly", targetDate, "savings")
	assert.Error(t, errSavings)
	assert.Contains(t, errSavings.Error(), "viewType 'savings' not yet implemented")

	_, errDebts := service.GetOrCreateFinancialSummary("monthly", targetDate, "debts")
	assert.Error(t, errDebts)
	assert.Contains(t, errDebts.Error(), "viewType 'debts' not yet implemented")
}

func TestGetOrCreateFinancialSummary_ViewInvalid(t *testing.T) {
	db := setupSummaryTestDB(t)
	service := NewSummaryService(db)
	targetDate := time.Now()

	_, err := service.GetOrCreateFinancialSummary("monthly", targetDate, "invalid_view")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid viewType 'invalid_view'")
}

// --- Tests for InvalidateSummariesForDate ---

func TestInvalidateSummariesForDate(t *testing.T) {
	itemDate := time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC)
	summaryTypesToInvalidate := []string{"monthly", "weekly", "yearly"}

	// Calculate expected period start dates
	monthlyStartDate, _, _ := CalculatePeriodDates(itemDate, "monthly")
	weeklyStartDate, _, _ := CalculatePeriodDates(itemDate, "weekly")
	yearlyStartDate, _, _ := CalculatePeriodDates(itemDate, "yearly")

	t.Run("InvalidateExistingSummaries", func(t *testing.T) {
		db := setupSummaryTestDB(t)
		service := NewSummaryService(db)

		// Seed summaries for the itemDate's periods
		seedFinancialSummary(t, db, "monthly", itemDate, 1000, 500)
		seedFinancialSummary(t, db, "weekly", itemDate, 200, 100)
		seedFinancialSummary(t, db, "yearly", itemDate, 5000, 2000)
		// Seed a summary for a different period that should NOT be deleted
		otherDate := itemDate.AddDate(0, -2, 0) // Two months before
		seedFinancialSummary(t, db, "monthly", otherDate, 100, 50)


		err := service.InvalidateSummariesForDate(itemDate, summaryTypesToInvalidate)
		assert.NoError(t, err)

		// Verify targeted summaries are deleted
		var count int64
		db.Model(&models.FinancialSummary{}).Where("summary_type = ? AND period_start_date = ?", "monthly", monthlyStartDate).Count(&count)
		assert.Equal(t, int64(0), count, "Monthly summary should be deleted")

		db.Model(&models.FinancialSummary{}).Where("summary_type = ? AND period_start_date = ?", "weekly", weeklyStartDate).Count(&count)
		assert.Equal(t, int64(0), count, "Weekly summary should be deleted")

		db.Model(&models.FinancialSummary{}).Where("summary_type = ? AND period_start_date = ?", "yearly", yearlyStartDate).Count(&count)
		assert.Equal(t, int64(0), count, "Yearly summary should be deleted")

		// Verify the other summary still exists
		otherMonthlyStartDate, _, _ := CalculatePeriodDates(otherDate, "monthly")
		db.Model(&models.FinancialSummary{}).Where("summary_type = ? AND period_start_date = ?", "monthly", otherMonthlyStartDate).Count(&count)
		assert.Equal(t, int64(1), count, "Other monthly summary should NOT be deleted")
	})

	t.Run("DateWithNoMatchingSummaries", func(t *testing.T) {
		db := setupSummaryTestDB(t)
		service := NewSummaryService(db)

		nonExistentItemDate := time.Date(2020, time.January, 1, 0,0,0,0, time.UTC)

		err := service.InvalidateSummariesForDate(nonExistentItemDate, summaryTypesToInvalidate)
		assert.NoError(t, err, "Should not error if no summaries found to delete")

		var count int64
		db.Model(&models.FinancialSummary{}).Count(&count)
		assert.Equal(t, int64(0), count, "No summaries should exist or be created")
	})

	t.Run("EmptySummaryPeriodTypesSlice", func(t *testing.T) {
		db := setupSummaryTestDB(t)
		service := NewSummaryService(db)

		seededSummary := seedFinancialSummary(t, db, "monthly", itemDate, 100, 50)

		err := service.InvalidateSummariesForDate(itemDate, []string{})
		assert.NoError(t, err)

		var count int64
		db.Model(&models.FinancialSummary{}).Where("id = ?", seededSummary.ID).Count(&count)
		assert.Equal(t, int64(1), count, "Summary should not be deleted for empty period types slice")
	})

	t.Run("SpecificPeriodTypeInvalidation", func(t *testing.T) {
		db := setupSummaryTestDB(t)
		service := NewSummaryService(db)

		seedFinancialSummary(t, db, "monthly", itemDate, 100, 10)
		seedFinancialSummary(t, db, "weekly", itemDate, 200, 20)
		seedFinancialSummary(t, db, "yearly", itemDate, 300, 30)

		err := service.InvalidateSummariesForDate(itemDate, []string{"monthly"})
		assert.NoError(t, err)

		var count int64
		db.Model(&models.FinancialSummary{}).Where("summary_type = ? AND period_start_date = ?", "monthly", monthlyStartDate).Count(&count)
		assert.Equal(t, int64(0), count, "Monthly summary should be deleted")

		db.Model(&models.FinancialSummary{}).Where("summary_type = ? AND period_start_date = ?", "weekly", weeklyStartDate).Count(&count)
		assert.Equal(t, int64(1), count, "Weekly summary should NOT be deleted")

		db.Model(&models.FinancialSummary{}).Where("summary_type = ? AND period_start_date = ?", "yearly", yearlyStartDate).Count(&count)
		assert.Equal(t, int64(1), count, "Yearly summary should NOT be deleted")
	})

	t.Run("ErrorInCalculatePeriodDates", func(t *testing.T) {
		db := setupSummaryTestDB(t)
		service := NewSummaryService(db)

		err := service.InvalidateSummariesForDate(itemDate, []string{"invalid-type"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to calculate period for invalid-type")
		assert.Contains(t, err.Error(), "invalid summary type: invalid-type")
	})
}
