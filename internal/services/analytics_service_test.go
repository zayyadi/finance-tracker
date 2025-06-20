package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zayyadi/finance-tracker/internal/database" // Corrected import path for CustomDate
	"github.com/zayyadi/finance-tracker/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB initializes an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err, "Failed to connect to in-memory SQLite")

	// Auto-migrate schemas.
	// GORM will create tables based on these structs. Since UserID is not an active field
	// in models.Income and models.Expense, the created tables for tests will not have user_id columns.
	// This aligns with testing the service logic which operates without user_id.

	// Drop tables first to ensure a clean state for each test
	// Order matters due to potential foreign key constraints if they were active
	// (though for our test setup with GORM creating schema from structs without UserID, FKs are not an issue)
	db.Exec("DROP TABLE IF EXISTS FinancialSummary")
	db.Exec("DROP TABLE IF EXISTS Debts") // Assuming Debt model might exist from other tests/full schema
	db.Exec("DROP TABLE IF EXISTS Savings")// Assuming Savings model might exist
	db.Exec("DROP TABLE IF EXISTS Expenses")
	db.Exec("DROP TABLE IF EXISTS Income")
	db.Exec("DROP TABLE IF EXISTS Users")


	err = db.AutoMigrate(&models.User{}, &models.Income{}, &models.Expense{}, &models.FinancialSummary{}, &models.Debt{}, &models.Savings{})
	assert.NoError(t, err, "Failed to auto-migrate models")

	return db
}

// seedExpenses populates the database with expense data.
func seedExpenses(t *testing.T, db *gorm.DB, expenses []models.Expense) {
	for _, expense := range expenses {
		err := db.Create(&expense).Error
		assert.NoError(t, err, "Failed to seed expense: %+v", expense)
	}
}

// seedIncomes populates the database with income data.
func seedIncomes(t *testing.T, db *gorm.DB, incomes []models.Income) {
	for _, income := range incomes {
		err := db.Create(&income).Error
		assert.NoError(t, err, "Failed to seed income: %+v", income)
	}
}

// TestGetExpenseBreakdownByCategory_NoData tests behavior when no expense data exists.
func TestGetExpenseBreakdownByCategory_NoData(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	analyticsService := NewAnalyticsService(db)

	targetDate := time.Date(2023, time.November, 15, 0, 0, 0, 0, time.UTC)
	stats, err := analyticsService.GetExpenseBreakdownByCategory(targetDate)

	assert.NoError(t, err)
	assert.Empty(t, stats, "Expected empty stats for no data")
}

// TestGetExpenseBreakdownByCategory_WithData tests expense breakdown with various data points.
func TestGetExpenseBreakdownByCategory_WithData(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	analyticsService := NewAnalyticsService(db)

	// Target month: November 2023
	nov2023 := time.Date(2023, time.November, 1, 0, 0, 0, 0, time.UTC)

	expensesToSeed := []models.Expense{
		// Current month (November 2023)
		{Amount: 50, Category: "Food", Date: database.CustomDate{Time: nov2023.AddDate(0, 0, 1)}},      // Nov 2
		{Amount: 30, Category: "Transport", Date: database.CustomDate{Time: nov2023.AddDate(0, 0, 2)}}, // Nov 3
		{Amount: 20, Category: "Food", Date: database.CustomDate{Time: nov2023.AddDate(0, 0, 3)}},      // Nov 4
		{Amount: 70, Category: "Utilities", Date: database.CustomDate{Time: nov2023.AddDate(0, 0, 5)}}, // Nov 6
		// Previous month (October 2023) - Should be ignored
		{Amount: 100, Category: "Shopping", Date: database.CustomDate{Time: nov2023.AddDate(0, -1, 1)}}, // Oct 2
		// Next month (December 2023) - Should be ignored
		{Amount: 60, Category: "Entertainment", Date: database.CustomDate{Time: nov2023.AddDate(0, 1, 1)}}, // Dec 2
	}
	seedExpenses(t, db, expensesToSeed)

	stats, err := analyticsService.GetExpenseBreakdownByCategory(nov2023.AddDate(0,0,14)) // Target date within Nov 2023
	assert.NoError(t, err)
	assert.Len(t, stats, 3, "Expected 3 categories for November 2023")

	// Results are ordered by total_amount DESC by the service
	// Expected: Food: 70, Utilities: 70, Transport: 30
	// Note: Order between Food and Utilities might vary if amounts are equal, which they are not here (Food 50+20=70)

	foundFood := false
	foundTransport := false
	foundUtilities := false

	for _, stat := range stats {
		if stat.Category == "Food" {
			assert.Equal(t, 70.0, stat.TotalAmount, "Food total amount incorrect")
			foundFood = true
		} else if stat.Category == "Transport" {
			assert.Equal(t, 30.0, stat.TotalAmount, "Transport total amount incorrect")
			foundTransport = true
		} else if stat.Category == "Utilities" {
			assert.Equal(t, 70.0, stat.TotalAmount, "Utilities total amount incorrect")
			foundUtilities = true
		}
	}
	assert.True(t, foundFood, "Food category not found")
	assert.True(t, foundTransport, "Transport category not found")
	assert.True(t, foundUtilities, "Utilities category not found")

	// Check order (Food and Utilities are 70, Transport is 30)
	// The service orders by "total_amount DESC". If amounts are equal, DB might not guarantee order
	// for items with the same total_amount.
	// So, we check that the items exist and have correct amounts, and that Transport is last.
	if len(stats) == 3 {
		// Check that Transport is the last one with 30.0
		assert.Equal(t, "Transport", stats[2].Category)
		assert.Equal(t, 30.0, stats[2].TotalAmount)

		// Check that the first two items are Food and Utilities, both with 70.0
		// Their relative order (stats[0] vs stats[1]) can vary.
		isFoodFirst := stats[0].Category == "Food" && stats[0].TotalAmount == 70.0 && stats[1].Category == "Utilities" && stats[1].TotalAmount == 70.0
		isUtilitiesFirst := stats[0].Category == "Utilities" && stats[0].TotalAmount == 70.0 && stats[1].Category == "Food" && stats[1].TotalAmount == 70.0

		assert.True(t, isFoodFirst || isUtilitiesFirst, "Expected Food and Utilities (both 70.0) to be the top two categories.")
	}
}


// TestGetIncomeExpenseTrend_NoData tests behavior with no income/expense data.
func TestGetIncomeExpenseTrend_NoData(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	analyticsService := NewAnalyticsService(db)

	numMonths := 3
	trend, err := analyticsService.GetIncomeExpenseTrend(numMonths)

	assert.NoError(t, err)
	assert.Len(t, trend, numMonths, "Expected N entries for N months")

	for i, monthlyStat := range trend {
		assert.Equal(t, 0.0, monthlyStat.TotalIncome, "Expected 0 income for month %d", i)
		assert.Equal(t, 0.0, monthlyStat.TotalExpenses, "Expected 0 expenses for month %d", i)
		// Check month format, e.g., "YYYY-MM"
		expectedMonth := time.Now().AddDate(0, -(numMonths-1-i), 0)
		assert.Equal(t, expectedMonth.Format("2006-01"), monthlyStat.Month, "Month format or value incorrect for month %d", i)
	}
}

// TestGetIncomeExpenseTrend_WithData tests income/expense trend with data.
func TestGetIncomeExpenseTrend_WithData(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	analyticsService := NewAnalyticsService(db)

	numMonths := 3
	now := time.Now().In(time.UTC) // Ensure UTC for consistency

	// Month 1 (2 months ago)
	month1 := now.AddDate(0, -2, 0)
	month1Start := time.Date(month1.Year(), month1.Month(), 1, 0, 0, 0, 0, time.UTC)
	incomesMonth1 := []models.Income{
		{Amount: 1000, Date: database.CustomDate{Time: month1Start.AddDate(0, 0, 1)}},
	}
	expensesMonth1 := []models.Expense{
		{Amount: 200, Category: "Groceries", Date: database.CustomDate{Time: month1Start.AddDate(0, 0, 2)}},
	}
	seedIncomes(t, db, incomesMonth1)
	seedExpenses(t, db, expensesMonth1)

	// Month 2 (1 month ago)
	month2 := now.AddDate(0, -1, 0)
	month2Start := time.Date(month2.Year(), month2.Month(), 1, 0, 0, 0, 0, time.UTC)
	incomesMonth2 := []models.Income{
		{Amount: 1200, Date: database.CustomDate{Time: month2Start.AddDate(0, 0, 1)}},
	}
	expensesMonth2 := []models.Expense{
		{Amount: 300, Category: "Utilities", Date: database.CustomDate{Time: month2Start.AddDate(0, 0, 2)}},
	}
	seedIncomes(t, db, incomesMonth2)
	seedExpenses(t, db, expensesMonth2)

	// Month 3 (current month)
	month3Start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	incomesMonth3 := []models.Income{
		{Amount: 1100, Date: database.CustomDate{Time: month3Start.AddDate(0, 0, 1)}},
	}
	expensesMonth3 := []models.Expense{
		{Amount: 250, Category: "Dining", Date: database.CustomDate{Time: month3Start.AddDate(0, 0, 2)}},
		// Add an expense from a different month to ensure it's not picked up for current month's expenses
		{Amount: 5000, Category: "Old", Date: database.CustomDate{Time: month1Start.AddDate(0,0,5)}},
	}
	seedIncomes(t, db, incomesMonth3)
	seedExpenses(t, db, expensesMonth3)


	trend, err := analyticsService.GetIncomeExpenseTrend(numMonths)
	assert.NoError(t, err)
	assert.Len(t, trend, numMonths, "Expected N trend entries for N months")

	// Trend should be chronological (oldest to newest)
	// Assert Month 1
	assert.Equal(t, month1Start.Format("2006-01"), trend[0].Month)
	assert.Equal(t, 1000.0, trend[0].TotalIncome)
	assert.Equal(t, 200.0 + 5000.0, trend[0].TotalExpenses) // 5000 was also in month1

	// Assert Month 2
	assert.Equal(t, month2Start.Format("2006-01"), trend[1].Month)
	assert.Equal(t, 1200.0, trend[1].TotalIncome)
	assert.Equal(t, 300.0, trend[1].TotalExpenses)

	// Assert Month 3 (current)
	assert.Equal(t, month3Start.Format("2006-01"), trend[2].Month)
	assert.Equal(t, 1100.0, trend[2].TotalIncome)
	assert.Equal(t, 250.0, trend[2].TotalExpenses)
}

func TestGetExpenseBreakdownByCategory_Order(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	analyticsService := NewAnalyticsService(db)

	targetMonth := time.Date(2024, time.January, 1, 0,0,0,0, time.UTC)

	expensesToSeed := []models.Expense{
		{Amount: 10, Category: "C", Date: database.CustomDate{Time: targetMonth.AddDate(0,0,1)}},
		{Amount: 30, Category: "A", Date: database.CustomDate{Time: targetMonth.AddDate(0,0,2)}},
		{Amount: 20, Category: "B", Date: database.CustomDate{Time: targetMonth.AddDate(0,0,3)}},
	}
	seedExpenses(t, db, expensesToSeed)

	stats, err := analyticsService.GetExpenseBreakdownByCategory(targetMonth)
	assert.NoError(t, err)
	assert.Len(t, stats, 3)

	assert.Equal(t, "A", stats[0].Category)
	assert.Equal(t, 30.0, stats[0].TotalAmount)

	assert.Equal(t, "B", stats[1].Category)
	assert.Equal(t, 20.0, stats[1].TotalAmount)

	assert.Equal(t, "C", stats[2].Category)
	assert.Equal(t, 10.0, stats[2].TotalAmount)
}

// TestGetIncomeExpenseTrend_WithData_EdgeCase_NoExpensesInOneMonth tests behavior
// when a month in the trend has income but no expenses.
func TestGetIncomeExpenseTrend_WithData_EdgeCase_NoExpensesInOneMonth(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	analyticsService := NewAnalyticsService(db)

	numMonths := 2
	now := time.Now().In(time.UTC)

	// Month 1 (1 month ago) - Income and Expenses
	month1 := now.AddDate(0, -1, 0)
	month1Start := time.Date(month1.Year(), month1.Month(), 1, 0, 0, 0, 0, time.UTC)
	seedIncomes(t, db, []models.Income{
		{Amount: 1000, Date: database.CustomDate{Time: month1Start.AddDate(0, 0, 5)}},
	})
	seedExpenses(t, db, []models.Expense{
		{Amount: 200, Category: "Food", Date: database.CustomDate{Time: month1Start.AddDate(0, 0, 5)}},
	})

	// Month 2 (current month) - Only Income, No Expenses
	month2Start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	seedIncomes(t, db, []models.Income{
		{Amount: 1500, Date: database.CustomDate{Time: month2Start.AddDate(0, 0, 5)}},
	})
	// No expenses seeded for month2

	trend, err := analyticsService.GetIncomeExpenseTrend(numMonths)
	assert.NoError(t, err)
	assert.Len(t, trend, numMonths)

	// Assert Month 1
	assert.Equal(t, month1Start.Format("2006-01"), trend[0].Month)
	assert.Equal(t, 1000.0, trend[0].TotalIncome)
	assert.Equal(t, 200.0, trend[0].TotalExpenses)

	// Assert Month 2
	assert.Equal(t, month2Start.Format("2006-01"), trend[1].Month)
	assert.Equal(t, 1500.0, trend[1].TotalIncome)
	assert.Equal(t, 0.0, trend[1].TotalExpenses, "Expected 0 expenses for current month")
}

// TestGetIncomeExpenseTrend_WithData_EdgeCase_NoIncomeInOneMonth tests behavior
// when a month in the trend has expenses but no income.
func TestGetIncomeExpenseTrend_WithData_EdgeCase_NoIncomeInOneMonth(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	analyticsService := NewAnalyticsService(db)

	numMonths := 2
	now := time.Now().In(time.UTC)

	// Month 1 (1 month ago) - Income and Expenses
	month1 := now.AddDate(0, -1, 0)
	month1Start := time.Date(month1.Year(), month1.Month(), 1, 0, 0, 0, 0, time.UTC)
	seedIncomes(t, db, []models.Income{
		{Amount: 1000, Date: database.CustomDate{Time: month1Start.AddDate(0, 0, 5)}},
	})
	seedExpenses(t, db, []models.Expense{
		{Amount: 200, Category: "Food", Date: database.CustomDate{Time: month1Start.AddDate(0, 0, 5)}},
	})

	// Month 2 (current month) - Only Expenses, No Income
	month2Start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	seedExpenses(t, db, []models.Expense{
		{Amount: 300, Category: "Utilities", Date: database.CustomDate{Time: month2Start.AddDate(0, 0, 5)}},
	})
	// No income seeded for month2

	trend, err := analyticsService.GetIncomeExpenseTrend(numMonths)
	assert.NoError(t, err)
	assert.Len(t, trend, numMonths)

	// Assert Month 1
	assert.Equal(t, month1Start.Format("2006-01"), trend[0].Month)
	assert.Equal(t, 1000.0, trend[0].TotalIncome)
	assert.Equal(t, 200.0, trend[0].TotalExpenses)

	// Assert Month 2
	assert.Equal(t, month2Start.Format("2006-01"), trend[1].Month)
	assert.Equal(t, 0.0, trend[1].TotalIncome, "Expected 0 income for current month")
	assert.Equal(t, 300.0, trend[1].TotalExpenses)
}
