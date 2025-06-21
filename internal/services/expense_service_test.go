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
	"gorm.io/gorm/logger" // Added for GORM logging
)

// setupExpenseTestDB initializes an in-memory SQLite database for expense service testing.
func setupExpenseTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	assert.NoError(t, err, "Failed to connect to in-memory SQLite")

	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })

	// Drop tables first for a clean state
	db.Exec("DROP TABLE IF EXISTS Expenses")
	db.Exec("DROP TABLE IF EXISTS Users") // In case of implicit dependencies

	// Auto-migrate schemas based on GORM structs.
	// Since UserID is not an active field in models.Expense, the created table will not have user_id.
	err = db.AutoMigrate(&models.User{}, &models.Expense{})
	assert.NoError(t, err, "Failed to auto-migrate models")

	// Optional: Create a dummy user if needed for any other service interactions not directly tested.
	// db.Create(&models.User{Username: "testuser", Email: "test@example.com", PasswordHash: "hash"})


	return db
}

// seedExpensesForTest populates the database with expense data for testing GetExpenses.
func seedExpensesForTest(t *testing.T, db *gorm.DB, expenses []models.Expense) {
	for _, expense := range expenses {
		err := db.Create(&expense).Error
		assert.NoError(t, err, "Failed to seed expense: %+v", expense)
	}
}

func TestGetExpenses_DateRange_NoData(t *testing.T) {
	db := setupExpenseTestDB(t)
	service := NewExpenseService(db)

	startDate := "2023-01-01"
	endDate := "2023-01-31"
	expenses, err := service.GetExpenses(0, 10, startDate, endDate)

	assert.NoError(t, err)
	assert.Empty(t, expenses)
}

func TestGetExpenses_DateRange_WithData_InRange(t *testing.T) {
	db := setupExpenseTestDB(t)
	service := NewExpenseService(db)

	expensesToSeed := []models.Expense{
		{Amount: 100, Category: "Food", Date: database.CustomDate{Time: time.Date(2023, time.January, 10, 0, 0, 0, 0, time.UTC)}},
		{Amount: 50, Category: "Transport", Date: database.CustomDate{Time: time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC)}},
		{Amount: 200, Category: "Shopping", Date: database.CustomDate{Time: time.Date(2023, time.February, 5, 0, 0, 0, 0, time.UTC)}}, // Outside range
	}
	seedExpensesForTest(t, db, expensesToSeed)

	startDate := "2023-01-01"
	endDate := "2023-01-31"
	expenses, err := service.GetExpenses(0, 10, startDate, endDate)

	assert.NoError(t, err)
	assert.Len(t, expenses, 2)
	assert.Equal(t, float64(50), expenses[0].Amount)    // Sorted by date desc
	assert.Equal(t, float64(100), expenses[1].Amount)
}

func TestGetExpenses_DateRange_WithData_OutsideRange(t *testing.T) {
	db := setupExpenseTestDB(t)
	service := NewExpenseService(db)

	expensesToSeed := []models.Expense{
		{Amount: 100, Category: "Food", Date: database.CustomDate{Time: time.Date(2023, time.February, 10, 0, 0, 0, 0, time.UTC)}},
	}
	seedExpensesForTest(t, db, expensesToSeed)

	startDate := "2023-01-01"
	endDate := "2023-01-31"
	expenses, err := service.GetExpenses(0, 10, startDate, endDate)

	assert.NoError(t, err)
	assert.Empty(t, expenses)
}

func TestGetExpenses_NoDateRange(t *testing.T) {
	db := setupExpenseTestDB(t)
	service := NewExpenseService(db)

	expensesToSeed := []models.Expense{
		{Amount: 100, Category: "Food", Date: database.CustomDate{Time: time.Date(2023, time.January, 10, 0, 0, 0, 0, time.UTC)}},
		{Amount: 50, Category: "Transport", Date: database.CustomDate{Time: time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC)}},
	}
	seedExpensesForTest(t, db, expensesToSeed)

	expenses, err := service.GetExpenses(0, 10, "", "") // No date range

	assert.NoError(t, err)
	assert.Len(t, expenses, 2) // Should return all, paginated
}

func TestGetExpenses_InvalidDateStrings(t *testing.T) {
	db := setupExpenseTestDB(t)
	service := NewExpenseService(db)

	_, err := service.GetExpenses(0, 10, "not-a-date", "2023-01-31")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid start date format")

	_, err = service.GetExpenses(0, 10, "2023-01-01", "also-not-a-date")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid end date format")
}

func TestGetExpenses_PaginationWithDateRange(t *testing.T) {
	db := setupExpenseTestDB(t)
	service := NewExpenseService(db)

	seedData := []models.Expense{
		{Amount: 10, Category: "Test", Date: database.CustomDate{Time: time.Date(2023, time.March, 1, 0, 0, 0, 0, time.UTC)}},
		{Amount: 20, Category: "Test", Date: database.CustomDate{Time: time.Date(2023, time.March, 2, 0, 0, 0, 0, time.UTC)}},
		{Amount: 30, Category: "Test", Date: database.CustomDate{Time: time.Date(2023, time.March, 3, 0, 0, 0, 0, time.UTC)}},
		{Amount: 40, Category: "Test", Date: database.CustomDate{Time: time.Date(2023, time.March, 4, 0, 0, 0, 0, time.UTC)}},
		{Amount: 50, Category: "Test", Date: database.CustomDate{Time: time.Date(2023, time.February, 5, 0, 0, 0, 0, time.UTC)}}, // Outside March
	}
	seedExpensesForTest(t, db, seedData)

	startDate := "2023-03-01"
	endDate := "2023-03-31"

	// Get first page
	expensesPage1, err1 := service.GetExpenses(0, 2, startDate, endDate)
	assert.NoError(t, err1)
	assert.Len(t, expensesPage1, 2)
	assert.Equal(t, float64(40), expensesPage1[0].Amount) // March 4 (latest due to Order("date desc"))
	assert.Equal(t, float64(30), expensesPage1[1].Amount) // March 3

	// Get second page
	// Enable GORM logging for this specific call
	originalLogger := db.Logger
	db.Logger = db.Logger.LogMode(logger.Info)

	expensesPage2, err2 := service.GetExpenses(2, 2, startDate, endDate)

	db.Logger = originalLogger // Restore original logger

	assert.NoError(t, err2)
	assert.Len(t, expensesPage2, 2)
	assert.Equal(t, float64(20), expensesPage2[0].Amount) // March 2
	assert.Equal(t, float64(10), expensesPage2[1].Amount) // March 1
}
