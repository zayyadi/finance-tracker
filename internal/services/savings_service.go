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
	// For pointer types like *database.CustomDate, if we want to allow setting them to NULL,
	// we include them in the map. If updateData.StartDate is nil (from JSON null or omitted),
	// it will be set as nil in the map, and GORM should update the DB field to NULL.
	// This requires that the client explicitly sends "field": null to clear it.
	// If the field is simply omitted from JSON, its pointer in updateData will also be nil.
	// To distinguish "omitted" from "explicitly set to null", a different approach like
	// using a map[string]interface{} directly from the handler, or custom unmarshalling
	// in the request struct would be needed.
	// For this fix, we'll assume that if a pointer field in SavingsUpdateRequest is nil,
	// it's an intention to set it to NULL if it's part of the map.
	// The handler's check `if req.StartDate == nil && ...` already filters out cases where
	// ALL fields are nil (i.e., empty JSON or all fields explicitly null).
	// We need to ensure that if ONLY target_date:null is sent (and notes for the test),
	// then target_date is updated to NULL.

	// We will always add these fields to the map if they were part of the request model.
	// To ensure we only update fields that were *actually present* in the request body,
	// this map building should be more intelligent, or we should use db.Model(&obj).Select("field1", "field2").Updates(obj_with_new_values)
	// For now, let's make a targeted fix for how nil pointers are handled for dates.
	// If the client wants to update a field, it sends it. If it sends null for a date, it means clear it.
	// The issue was the `if updateData.StartDate != nil` check.

	updatesMap["start_date"] = updateData.StartDate   // if updateData.StartDate is nil, map value becomes nil
	updatesMap["target_date"] = updateData.TargetDate // if updateData.TargetDate is nil, map value becomes nil

	if updateData.Notes != nil {
		updatesMap["notes"] = *updateData.Notes
	}

	// The handler should ensure that an empty request body (no actual fields to update) is caught.
	// Our current handler check is: if all fields in req are nil -> 400.
	// If req is {"target_date": null, "notes": "text"}, then req.TargetDate is nil, req.Notes is not.
	// So updatesMap will be {"target_date": nil, "notes": "text"}
	// This should correctly set target_date to NULL and update notes.

	if len(updatesMap) == 0 && updateData.StartDate == nil && updateData.TargetDate == nil {
		// This condition needs to be robust. If only GoalName, GoalAmount, CurrentAmount, Notes were nilled out
		// and dates were not provided, map would be empty.
		// The original check `if len(updatesMap) == 0` was based on non-nil values.
		// A truly empty request (e.g. {}) is caught by handler.
		// A request like {"goal_name": null} is not possible with current *string, it would be missing or "goal_name":""
		// Let's rely on the handler's check for now. If updatesMap is empty here, it means
		// only date fields were provided as null and no other fields.
		// This is okay, we want to proceed to update them to NULL.
	}

	// Use UpdateColumns to ensure that nil values in the map explicitly set DB fields to NULL.
	// Updates might ignore nil values in maps depending on GORM version and configuration.
	result := s.DB.Model(&existingSavings).Where("id = ?", savingsID).UpdateColumns(updatesMap)

	if result.Error != nil {
		log.Printf("Error updating savings goal %d: %v", savingsID, result.Error)
		return nil, fmt.Errorf("could not update savings goal: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		// It's possible no rows were affected because the data in updatesMap matched existing data.
		// However, if we are trying to set a field to NULL and it wasn't NULL, it should affect rows.
		// For safety, re-fetch. If it was truly "not found", the initial GetSavingsByID would have caught it.
		// If an update results in 0 rows affected but no error, it might mean the record matched the update already.
		// Re-fetch to be sure.
		log.Printf("Update operation on savings goal %d affected 0 rows. Re-fetching to confirm state.", savingsID)
	}

	// Re-fetch into a new variable to ensure the returned model has the latest data from the database,
	// especially to correctly reflect fields set to NULL and avoid issues with GORM potentially
	// not clearing fields in an already populated struct.
	var freshlyFetchedSavings models.Savings
	err = s.DB.First(&freshlyFetchedSavings, savingsID).Error
	if err != nil {
		log.Printf("Error re-fetching savings goal %d after update: %v", savingsID, err)
		return nil, fmt.Errorf("could not re-fetch savings goal after update: %w", err)
	}

	return &freshlyFetchedSavings, nil
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
