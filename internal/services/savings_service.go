package services

import (
	"errors"
	"fmt"
	"log"

	// "time" // Not needed directly if GORM handles timestamps in gorm.Model

	"github.com/zayyadi/finance-tracker/internal/database"
	"github.com/zayyadi/finance-tracker/internal/models"
	"gorm.io/gorm"
)

// SavingsService provides methods for managing savings goal records using GORM.
type SavingsService struct {
	DB *gorm.DB
}

// NewSavingsService creates a new SavingsService with a GORM database connection.
func NewSavingsService(db *gorm.DB) *SavingsService {
	if db == nil {
		log.Println("Warning: NewSavingsService called with nil DB, attempting to use global GetDB()")
		db = database.GetDB()
	}
	return &SavingsService{DB: db}
}

// CreateSavings inserts a new savings goal record.
func (s *SavingsService) CreateSavings(savings *models.Savings) error {
	if s.DB == nil {
		return fmt.Errorf("database connection not initialized in SavingsService")
	}
	result := s.DB.Create(savings)
	if result.Error != nil {
		// log.Printf("Error creating savings goal for user %d: %v", savings.UserID, result.Error) // UserID removed
		log.Printf("Error creating savings goal: %v", result.Error)
		return fmt.Errorf("could not create savings goal: %w", result.Error)
	}
	return nil
}

// GetSavingsByID retrieves a specific savings goal by its ID.
func (s *SavingsService) GetSavingsByID(savingsID uint) (*models.Savings, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in SavingsService")
	}
	var savings models.Savings
	result := s.DB.Where("id = ?", savingsID).First(&savings) // Removed userID condition
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("savings goal not found") // Simplified error
		}
		log.Printf("Error retrieving savings goal %d: %v", savingsID, result.Error)
		return nil, fmt.Errorf("could not retrieve savings goal: %w", result.Error)
	}
	return &savings, nil
}

// GetSavings retrieves all savings goals with pagination.
func (s *SavingsService) GetSavings(offset int, limit int) ([]models.Savings, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in SavingsService")
	}
	var savingsList []models.Savings
	result := s.DB.Offset(offset).Limit(limit). // Removed userID condition
							Order("target_date asc, created_at desc").
							Find(&savingsList)

	if result.Error != nil {
		log.Printf("Error retrieving savings goals: %v", result.Error)
		return nil, fmt.Errorf("could not retrieve savings goals: %w", result.Error)
	}
	if savingsList == nil {
		return []models.Savings{}, nil
	}
	return savingsList, nil
}

// UpdateSavings updates an existing savings goal.
func (s *SavingsService) UpdateSavings(savingsID uint, updateData *models.SavingsUpdateRequest) (*models.Savings, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection not initialized in SavingsService")
	}

	existingSavings, err := s.GetSavingsByID(savingsID) // Removed userID
	if err != nil {
		return nil, err
	}

	// and to correctly handle zero values if not using pointers for all fields in updateData.
	// For GORM, Model(&existingSavings).Updates(updateData) where updateData is a struct with pointer fields
	// (or gorm.Association for relationships) is often preferred.
	// If updateData struct has non-pointer fields, GORM will update them even if they are zero-valued in the request.
	// Using map or .Select for specific fields is safer for partial updates.

	updatesMap := make(map[string]interface{})
	if updateData.GoalName != nil {
		updatesMap["goal_name"] = *updateData.GoalName
	}
	if updateData.GoalAmount != nil {
		updatesMap["goal_amount"] = *updateData.GoalAmount
	}
	if updateData.CurrentAmount != nil {
		updatesMap["current_amount"] = *updateData.CurrentAmount
	}
	if updateData.StartDate != nil { // Assuming updateData.StartDate is *time.Time
		updatesMap["start_date"] = updateData.StartDate // Can be nil to set to NULL
	}
	if updateData.TargetDate != nil { // Assuming updateData.TargetDate is *time.Time
		updatesMap["target_date"] = updateData.TargetDate // Can be nil to set to NULL
	}
	if updateData.Notes != nil {
		updatesMap["notes"] = *updateData.Notes
	}

	if len(updatesMap) == 0 {
		return existingSavings, nil // No fields to update
	}

	result := s.DB.Model(&existingSavings).Where("id = ?", savingsID).Updates(updatesMap) // Removed userID

	if result.Error != nil {
		log.Printf("Error updating savings goal %d: %v", savingsID, result.Error)
		return nil, fmt.Errorf("could not update savings goal: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("savings goal not found during update (or no changes made)")
	}

	return existingSavings, nil
}

// DeleteSavings deletes a savings goal.
func (s *SavingsService) DeleteSavings(savingsID uint) error {
	if s.DB == nil {
		return fmt.Errorf("database connection not initialized in SavingsService")
	}
	result := s.DB.Where("id = ?", savingsID).Delete(&models.Savings{}) // Removed userID
	if result.Error != nil {
		log.Printf("Error deleting savings goal %d: %v", savingsID, result.Error)
		return fmt.Errorf("could not delete savings goal: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("savings goal not found, no rows deleted")
	}
	return nil
}
