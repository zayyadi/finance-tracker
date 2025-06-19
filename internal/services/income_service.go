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

// IncomeService provides methods for managing income records using GORM.
type IncomeService struct {
	DB *gorm.DB
}

// NewIncomeService creates a new IncomeService with a GORM database connection.
func NewIncomeService(db *gorm.DB) *IncomeService {
	if db == nil {
		db = database.GetDB() // Fallback
	}
	return &IncomeService{DB: db}
}

// CreateIncome inserts a new income record.
func (s *IncomeService) CreateIncome(income *models.Income) error {
	if s.DB == nil {
		return fmt.Errorf("database connection not initialized in IncomeService")
	}
	result := s.DB.Create(income)
	if result.Error != nil {
		// log.Printf("Error creating income for user %d: %v", income.UserID, result.Error) // UserID removed
		log.Printf("Error creating income: %v", result.Error)
		return fmt.Errorf("could not create income: %w", result.Error)
	}
	return nil
}

// GetIncomeByID retrieves a specific income record by its ID.
func (s *IncomeService) GetIncomeByID(incomeID uint) (*models.Income, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in IncomeService")
	}
	var income models.Income
	result := s.DB.Where("id = ?", incomeID).First(&income) // Removed userID condition
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("income record not found") // Simplified error message
		}
		log.Printf("Error retrieving income %d: %v", incomeID, result.Error)
		return nil, fmt.Errorf("could not retrieve income: %w", result.Error)
	}
	return &income, nil
}

// GetIncomes retrieves all income records with pagination.
func (s *IncomeService) GetIncomes(offset int, limit int) ([]models.Income, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in IncomeService")
	}
	var incomes []models.Income
	result := s.DB.Offset(offset).Limit(limit). // Removed userID condition
							Order("date desc, created_at desc").
							Find(&incomes)

	if result.Error != nil {
		log.Printf("Error retrieving incomes: %v", result.Error)
		return nil, fmt.Errorf("could not retrieve incomes: %w", result.Error)
	}
	if incomes == nil {
		return []models.Income{}, nil
	}
	return incomes, nil
}

// GetIncomesByDateRange retrieves all income records within a specific date range.
func (s *IncomeService) GetIncomesByDateRange(startDate, endDate time.Time) ([]models.Income, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in IncomeService")
	}
	var incomes []models.Income
	result := s.DB.Where("date BETWEEN ? AND ?", startDate, endDate). // Removed userID condition
										Order("date desc, created_at desc").
										Find(&incomes)

	if result.Error != nil {
		log.Printf("Error retrieving incomes by date range: %v", result.Error)
		return nil, fmt.Errorf("could not retrieve incomes by date range: %w", result.Error)
	}
	if incomes == nil {
		return []models.Income{}, nil
	}
	return incomes, nil
}

// UpdateIncome updates an existing income record.
func (s *IncomeService) UpdateIncome(incomeID uint, updateData *models.IncomeUpdateRequest) (*models.Income, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in IncomeService")
	}

	existingIncome, err := s.GetIncomeByID(incomeID) // Removed userID
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
	if updateData.Note != nil { // Note can be updated to an empty string
		updates["note"] = *updateData.Note
	}

	if len(updates) == 0 {
		// No actual fields to update, just return the existing record
		// or you could return an error indicating no update data was provided.
		return existingIncome, nil
	}

	result := s.DB.Model(&existingIncome).Where("id = ?", incomeID).Updates(updates) // Removed userID condition
	if result.Error != nil {
		log.Printf("Error updating income %d: %v", incomeID, result.Error)
		return nil, fmt.Errorf("could not update income: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("income record not found during update (or no changes made)")
	}

	return existingIncome, nil
}

// DeleteIncome deletes an income record.
func (s *IncomeService) DeleteIncome(incomeID uint) error {
	if s.DB == nil {
		return fmt.Errorf("database connection not initialized in IncomeService")
	}
	result := s.DB.Where("id = ?", incomeID).Delete(&models.Income{}) // Removed userID condition
	if result.Error != nil {
		log.Printf("Error deleting income %d: %v", incomeID, result.Error)
		return fmt.Errorf("could not delete income: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("income record not found, no rows deleted")
	}
	return nil
}
