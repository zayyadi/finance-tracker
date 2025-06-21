package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/zayyadi/finance-tracker/internal/database"
	"github.com/zayyadi/finance-tracker/internal/models"
	"github.com/zayyadi/finance-tracker/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupSavingsTestRouter initializes an in-memory SQLite database and sets up the Gin router
// with savings routes for testing.
func setupSavingsTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err, "Failed to connect to in-memory SQLite")

	sqlDB, err := db.DB()
	assert.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	// Drop tables first for a clean state
	db.Exec("DROP TABLE IF EXISTS Savings")
	db.Exec("DROP TABLE IF EXISTS Users") // In case of implicit dependencies or future use

	// Auto-migrate schemas
	err = db.AutoMigrate(&models.User{}, &models.Savings{})
	assert.NoError(t, err, "Failed to auto-migrate models")

	// Create a dummy user if any FK constraints might apply implicitly or for other services
	// For current Savings model, UserID is commented out, so direct FK is not an issue.
	// However, creating one is good practice for a more complete test setup.
	db.Create(&models.User{Username: "testuser", Email: "test@example.com", PasswordHash: "hash"})


	savingsService := services.NewSavingsService(db)
	savingsHandler := NewSavingsHandler(savingsService)

	router := gin.Default()
	// Register only the routes needed for these tests
	router.POST("/savings", savingsHandler.CreateSavingsHandler)
	router.PUT("/savings/:id", savingsHandler.UpdateSavingsHandler)
	// Add GET /savings/:id if needed for easy verification, though direct DB check is also good
	router.GET("/savings/:id", savingsHandler.GetSavingsHandler)


	return router, db
}

func TestCreateSavingsHandler_WithCustomDateFormat(t *testing.T) {
	router, db := setupSavingsTestRouter(t)

	targetDateStr := "2025-07-15"
	payload := fmt.Sprintf(`{
		"goal_name": "Test Vacation",
		"goal_amount": 1200.50,
		"target_date": "%s"
	}`, targetDateStr)

	req, _ := http.NewRequest("POST", "/savings", bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "HTTP status code mismatch")

	var createdGoal models.Savings
	err := json.Unmarshal(rr.Body.Bytes(), &createdGoal)
	assert.NoError(t, err, "Failed to unmarshal response body")

	assert.NotNil(t, createdGoal.TargetDate, "TargetDate should not be nil")
	if createdGoal.TargetDate != nil {
		expectedTime, _ := time.Parse("2006-01-02", targetDateStr)
		assert.Equal(t, expectedTime.Year(), createdGoal.TargetDate.Time.Year(), "Year mismatch")
		assert.Equal(t, expectedTime.Month(), createdGoal.TargetDate.Time.Month(), "Month mismatch")
		assert.Equal(t, expectedTime.Day(), createdGoal.TargetDate.Time.Day(), "Day mismatch")
	}

	// Verify in DB
	var dbGoal models.Savings
	result := db.First(&dbGoal, createdGoal.ID)
	assert.NoError(t, result.Error, "Failed to fetch goal from DB")
	assert.NotNil(t, dbGoal.TargetDate, "DB TargetDate should not be nil")
	if dbGoal.TargetDate != nil {
		expectedTime, _ := time.Parse("2006-01-02", targetDateStr)
		assert.Equal(t, expectedTime.Year(), dbGoal.TargetDate.Time.Year(), "DB Year mismatch")
		assert.Equal(t, expectedTime.Month(), dbGoal.TargetDate.Time.Month(), "DB Month mismatch")
		assert.Equal(t, expectedTime.Day(), dbGoal.TargetDate.Time.Day(), "DB Day mismatch")
	}
	assert.Equal(t, "Test Vacation", dbGoal.GoalName)
	assert.Equal(t, 1200.50, dbGoal.GoalAmount)
}

func TestUpdateSavingsHandler_WithCustomDateFormat(t *testing.T) {
	router, db := setupSavingsTestRouter(t)

	// 1. Seed an initial savings goal
	initialTargetDateStr := "2024-12-01"
	initialTargetCustomDate, _ := time.Parse("2006-01-02", initialTargetDateStr)

	initialGoal := models.Savings{
		GoalName:   "Initial Goal",
		GoalAmount: 500.00,
		TargetDate: &database.CustomDate{Time: initialTargetCustomDate},
	}
	db.Create(&initialGoal)
	assert.NotZero(t, initialGoal.ID, "Failed to create initial savings goal")

	// 2. Prepare update request
	newTargetDateStr := "2026-08-20"
	updatePayload := fmt.Sprintf(`{
		"target_date": "%s",
		"goal_name": "Updated Holiday Plan"
	}`, newTargetDateStr)

	req, _ := http.NewRequest("PUT", "/savings/"+strconv.Itoa(int(initialGoal.ID)), bytes.NewBuffer([]byte(updatePayload)))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "HTTP status code mismatch for update")

	var updatedGoalResp models.Savings
	err := json.Unmarshal(rr.Body.Bytes(), &updatedGoalResp)
	assert.NoError(t, err, "Failed to unmarshal update response body")

	assert.NotNil(t, updatedGoalResp.TargetDate, "Updated TargetDate should not be nil in response")
	if updatedGoalResp.TargetDate != nil {
		expectedNewTime, _ := time.Parse("2006-01-02", newTargetDateStr)
		assert.Equal(t, expectedNewTime.Year(), updatedGoalResp.TargetDate.Time.Year(), "Updated Year mismatch in response")
		assert.Equal(t, expectedNewTime.Month(), updatedGoalResp.TargetDate.Time.Month(), "Updated Month mismatch in response")
		assert.Equal(t, expectedNewTime.Day(), updatedGoalResp.TargetDate.Time.Day(), "Updated Day mismatch in response")
	}
	assert.Equal(t, "Updated Holiday Plan", updatedGoalResp.GoalName, "GoalName mismatch in response")


	// 3. Verify in DB
	var dbGoal models.Savings
	result := db.First(&dbGoal, initialGoal.ID)
	assert.NoError(t, result.Error, "Failed to fetch updated goal from DB")

	assert.NotNil(t, dbGoal.TargetDate, "DB TargetDate should not be nil after update")
	if dbGoal.TargetDate != nil {
		expectedNewTime, _ := time.Parse("2006-01-02", newTargetDateStr)
		assert.Equal(t, expectedNewTime.Year(), dbGoal.TargetDate.Time.Year(), "DB Year mismatch after update")
		assert.Equal(t, expectedNewTime.Month(), dbGoal.TargetDate.Time.Month(), "DB Month mismatch after update")
		assert.Equal(t, expectedNewTime.Day(), dbGoal.TargetDate.Time.Day(), "DB Day mismatch after update")
	}
	assert.Equal(t, "Updated Holiday Plan", dbGoal.GoalName, "DB GoalName mismatch after update")
}

func TestCreateSavingsHandler_WithNullStartDate(t *testing.T) {
	router, db := setupSavingsTestRouter(t)

	targetDateStr := "2025-09-20"
	payload := fmt.Sprintf(`{
        "goal_name": "Future Purchase",
        "goal_amount": 300.75,
        "current_amount": 50.0,
        "start_date": null,
        "target_date": "%s",
        "notes": "Saving for something nice"
    }`, targetDateStr)

	req, _ := http.NewRequest("POST", "/savings", bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var createdGoal models.Savings
	err := json.Unmarshal(rr.Body.Bytes(), &createdGoal)
	assert.NoError(t, err)

	assert.Nil(t, createdGoal.StartDate, "StartDate should be nil in response")
	assert.NotNil(t, createdGoal.TargetDate)
	if createdGoal.TargetDate != nil {
		expectedTargetTime, _ := time.Parse("2006-01-02", targetDateStr)
		assert.Equal(t, expectedTargetTime.Day(), createdGoal.TargetDate.Time.Day())
	}


	// Verify in DB
	var dbGoal models.Savings
	db.First(&dbGoal, createdGoal.ID)
	assert.Nil(t, dbGoal.StartDate, "StartDate should be nil in DB")
	assert.NotNil(t, dbGoal.TargetDate)
}

func TestUpdateSavingsHandler_SetDateToNull(t *testing.T) {
	router, db := setupSavingsTestRouter(t)

    // Seed an initial savings goal with a target date
	initialTargetDate, _ := time.Parse("2006-01-02", "2025-01-01")
    initialGoal := models.Savings{
        GoalName:   "Goal to make TargetDate null",
        GoalAmount: 100.0,
        TargetDate: &database.CustomDate{Time: initialTargetDate},
    }
    db.Create(&initialGoal)
    assert.NotZero(t, initialGoal.ID)

    // Update TargetDate to null, and add another field to bypass "at least one field" check
    updatePayload := `{ "target_date": null, "notes": "Set target date to null" }`
    req, _ := http.NewRequest("PUT", "/savings/"+strconv.Itoa(int(initialGoal.ID)), bytes.NewBuffer([]byte(updatePayload)))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
    var updatedGoalResp models.Savings
    json.Unmarshal(rr.Body.Bytes(), &updatedGoalResp)
    // This assertion might fail if GORM re-fetches and Scan results in a non-nil CustomDate with zero Time

    // Verify in DB first - this is the most critical check
    var dbGoal models.Savings
    db.First(&dbGoal, initialGoal.ID)
    assert.Nil(t, dbGoal.TargetDate, "TargetDate should be nil in DB after update to null")

    // Then check response - if DB is nil, response should also be nil (or represent a zero time)
    // After JSON unmarshalling `null` into a *database.CustomDate, the pointer might be non-nil
    // but point to a CustomDate with a zero Time.
    assert.True(t, updatedGoalResp.TargetDate == nil || updatedGoalResp.TargetDate.Time.IsZero(), "TargetDate should be effectively nil in response after update to null")
}
