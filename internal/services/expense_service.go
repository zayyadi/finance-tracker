package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/zayyadi/finance-tracker/internal/database"
	"github.com/zayyadi/finance-tracker/internal/models"
	"gorm.io/gorm"
)

// ExpenseService provides methods for managing expense records using GORM.
type ExpenseService struct {
	DB *gorm.DB
}

// NewExpenseService creates a new ExpenseService with a GORM database connection.
func NewExpenseService(db *gorm.DB) *ExpenseService {
	if db == nil {
		log.Println("Warning: NewExpenseService called with nil DB, attempting to use global GetDB()")
		db = database.GetDB()
	}
	return &ExpenseService{DB: db}
}

// CreateExpense inserts a new expense record.
func (s *ExpenseService) CreateExpense(expense *models.Expense) error {
	if s.DB == nil {
		return fmt.Errorf("database connection not initialized in ExpenseService")
	}
	result := s.DB.Create(expense)
	if result.Error != nil {
		// log.Printf("Error creating expense for user %d: %v", expense.UserID, result.Error) // UserID removed
		log.Printf("Error creating expense: %v", result.Error)
		return fmt.Errorf("could not create expense: %w", result.Error)
	}
	return nil
}

// GetExpenseByID retrieves a specific expense record by its ID.
func (s *ExpenseService) GetExpenseByID(expenseID uint) (*models.Expense, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in ExpenseService")
	}
	var expense models.Expense
	result := s.DB.Where("id = ?", expenseID).First(&expense) // Removed userID condition
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("expense record not found") // Simplified error
		}
		log.Printf("Error retrieving expense %d: %v", expenseID, result.Error)
		return nil, fmt.Errorf("could not retrieve expense: %w", result.Error)
	}
	return &expense, nil
}

// GetExpenses retrieves all expense records with pagination.
func (s *ExpenseService) GetExpenses(offset int, limit int) ([]models.Expense, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in ExpenseService")
	}
	var expenses []models.Expense
	result := s.DB.Offset(offset).Limit(limit). // Removed userID condition
							Order("date desc, created_at desc").
							Find(&expenses)

	if result.Error != nil {
		log.Printf("Error retrieving expenses: %v", result.Error)
		return nil, fmt.Errorf("could not retrieve expenses: %w", result.Error)
	}
	if expenses == nil {
		return []models.Expense{}, nil
	}
	return expenses, nil
}

// GetExpensesByDateRange retrieves all expense records within a specific date range.
func (s *ExpenseService) GetExpensesByDateRange(startDate, endDate time.Time) ([]models.Expense, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in ExpenseService")
	}
	var expenses []models.Expense
	result := s.DB.Where("date BETWEEN ? AND ?", startDate, endDate). // Removed userID condition
										Order("date desc, created_at desc").
										Find(&expenses)

	if result.Error != nil {
		log.Printf("Error retrieving expenses by date range: %v", result.Error)
		return nil, fmt.Errorf("could not retrieve expenses by date range: %w", result.Error)
	}
	if expenses == nil {
		return []models.Expense{}, nil
	}
	return expenses, nil
}

// UpdateExpense updates an existing expense record.
func (s *ExpenseService) UpdateExpense(expenseID uint, updateData *models.ExpenseUpdateRequest) (*models.Expense, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in ExpenseService")
	}

	existingExpense, err := s.GetExpenseByID(expenseID) // Removed userID
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if updateData.Amount != nil {
		updates["amount"] = *updateData.Amount
	}
	if updateData.Category != nil {
		updates["category"] = *updateData.Category
	}
	if updateData.Date != nil {
		updates["date"] = *updateData.Date
	}
	if updateData.Note != nil {
		updates["note"] = *updateData.Note
	}

	if len(updates) == 0 {
		return existingExpense, nil
	}

	result := s.DB.Model(&existingExpense).Where("id = ?", expenseID).Updates(updates) // Removed userID condition
	if result.Error != nil {
		log.Printf("Error updating expense %d: %v", expenseID, result.Error)
		return nil, fmt.Errorf("could not update expense: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("expense record not found during update (or no changes made)")
	}
	return existingExpense, nil
}

// DeleteExpense deletes an expense record.
func (s *ExpenseService) DeleteExpense(expenseID uint) error {
	if s.DB == nil {
		return fmt.Errorf("database connection not initialized in ExpenseService")
	}
	result := s.DB.Where("id = ?", expenseID).Delete(&models.Expense{}) // Removed userID condition
	if result.Error != nil {
		log.Printf("Error deleting expense %d: %v", expenseID, result.Error)
		return fmt.Errorf("could not delete expense: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("expense record not found, no rows deleted")
	}
	return nil
}
