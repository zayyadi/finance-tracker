package services

import (
	"errors"
	"fmt"
	"log"

	// "time" // Not needed for GORM model timestamps if gorm.Model is used

	"github.com/zayyadi/finance-tracker/internal/database"
	"github.com/zayyadi/finance-tracker/internal/models"
	"gorm.io/gorm"
)

// DebtService provides methods for managing debt records using GORM.
type DebtService struct {
	DB *gorm.DB
}

// NewDebtService creates a new DebtService with a GORM database connection.
func NewDebtService(db *gorm.DB) *DebtService {
	if db == nil {
		log.Println("Warning: NewDebtService called with nil DB, attempting to use global GetDB()")
		db = database.GetDB()
	}
	return &DebtService{DB: db}
}

// CreateDebt inserts a new debt record.
func (s *DebtService) CreateDebt(debt *models.Debt) error {
	if s.DB == nil {
		return fmt.Errorf("database connection not initialized in DebtService")
	}
	result := s.DB.Create(debt)
	if result.Error != nil {
		// log.Printf("Error creating debt for user %d: %v", debt.UserID, result.Error) // UserID removed
		log.Printf("Error creating debt: %v", result.Error)
		return fmt.Errorf("could not create debt: %w", result.Error)
	}
	return nil
}

// GetDebtByID retrieves a specific debt record by its ID.
func (s *DebtService) GetDebtByID(debtID uint) (*models.Debt, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in DebtService")
	}
	var debt models.Debt
	result := s.DB.Where("id = ?", debtID).First(&debt) // Removed userID
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("debt record not found") // Simplified error
		}
		log.Printf("Error retrieving debt %d: %v", debtID, result.Error)
		return nil, fmt.Errorf("could not retrieve debt: %w", result.Error)
	}
	return &debt, nil
}

// GetDebts retrieves all debt records with pagination and optional status filter.
func (s *DebtService) GetDebts(offset int, limit int, statusFilter string) ([]models.Debt, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in DebtService")
	}

	query := s.DB.Model(&models.Debt{}) // Start query on the Debt model
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	var debts []models.Debt
	result := query.Offset(offset).Limit(limit).Order("due_date asc, created_at desc").Find(&debts) // Removed userID

	if result.Error != nil {
		log.Printf("Error retrieving debts (status: '%s'): %v", statusFilter, result.Error)
		return nil, fmt.Errorf("could not retrieve debts: %w", result.Error)
	}
	if debts == nil {
		return []models.Debt{}, nil
	}
	return debts, nil
}

// UpdateDebt updates an existing debt record.
func (s *DebtService) UpdateDebt(debtID uint, updateData *models.DebtUpdateRequest) (*models.Debt, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in DebtService")
	}

	existingDebt, err := s.GetDebtByID(debtID) // Removed userID
	if err != nil {
		return nil, err
	}

	updatesMap := make(map[string]interface{})
	if updateData.DebtorName != nil {
		updatesMap["debtor_name"] = *updateData.DebtorName
	}
	if updateData.Description != nil { // Allows setting description to empty string if desired
		updatesMap["description"] = *updateData.Description
	}
	if updateData.Amount != nil {
		updatesMap["amount"] = *updateData.Amount
	}
	if updateData.DueDate != nil {
		updatesMap["due_date"] = *updateData.DueDate
	}
	if updateData.Status != nil && *updateData.Status != "" {
		updatesMap["status"] = *updateData.Status
	}

	if len(updatesMap) == 0 {
		return existingDebt, nil // No fields to update
	}

	result := s.DB.Model(&existingDebt).Where("id = ?", debtID).Updates(updatesMap) // Removed userID
	if result.Error != nil {
		log.Printf("Error updating debt %d: %v", debtID, result.Error)
		return nil, fmt.Errorf("could not update debt: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("debt record not found during update (or no changes made)")
	}
	return existingDebt, nil
}

// DeleteDebt deletes a debt record.
func (s *DebtService) DeleteDebt(debtID uint) error {
	if s.DB == nil {
		return fmt.Errorf("database connection not initialized in DebtService")
	}
	result := s.DB.Where("id = ?", debtID).Delete(&models.Debt{}) // Removed userID
	if result.Error != nil {
		log.Printf("Error deleting debt %d: %v", debtID, result.Error)
		return fmt.Errorf("could not delete debt: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("debt record not found, no rows deleted")
	}
	return nil
}
