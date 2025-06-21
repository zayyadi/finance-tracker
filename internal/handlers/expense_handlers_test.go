package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time" // Added import for time

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/zayyadi/finance-tracker/internal/database" // Added import for database.CustomDate
	"github.com/zayyadi/finance-tracker/internal/models"
	"github.com/zayyadi/finance-tracker/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupExpenseTestRouter initializes an in-memory SQLite database and sets up the Gin router
// with expense routes for testing.
func setupExpenseTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// Use a unique DSN for each test to ensure isolation with cache=shared
	dsn := fmt.Sprintf("file:exp_handler_%s_%d?mode=memory&cache=shared", t.Name(), time.Now().UnixNano()) // Use NanoTime for uniqueness
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err, "Failed to connect to in-memory SQLite")

	sqlDB, errDB := db.DB()
	assert.NoError(t, errDB)
	t.Cleanup(func() { sqlDB.Close() })

	// Drop tables first for a clean state
	db.Exec("DROP TABLE IF EXISTS Expenses")
	db.Exec("DROP TABLE IF EXISTS Users")

	// Auto-migrate schemas
	err = db.AutoMigrate(&models.User{}, &models.Expense{})
	assert.NoError(t, err, "Failed to auto-migrate models")

	// Seed a dummy user (optional, but good practice if any underlying service logic might require it)
	db.Create(&models.User{Username: "testuser", Email: "test@example.com", PasswordHash: "hash"})


	expenseService := services.NewExpenseService(db)
	expenseHandler := NewExpenseHandler(expenseService)

	router := gin.Default()
	// Register expense routes
	router.POST("/expenses", expenseHandler.CreateExpenseHandler) // Needed for creating items if tests require it
	router.GET("/expenses/:id", expenseHandler.GetExpenseHandler)
	router.PUT("/expenses/:id", expenseHandler.UpdateExpenseHandler)
	router.DELETE("/expenses/:id", expenseHandler.DeleteExpenseHandler)
	router.GET("/expenses", expenseHandler.ListExpensesHandler)


	return router, db
}

var invalidIDs = []string{"not-a-number", " ", "1.0", "1a2b", "-1", "0"} // Removed "" as it tests router more than handler parsing for :id

func TestGetExpenseHandler_InvalidIDFormat(t *testing.T) {
	router, _ := setupExpenseTestRouter(t) // DB instance not strictly needed for this test

	for _, invalidID := range invalidIDs {
		t.Run(fmt.Sprintf("ID_%s", invalidID), func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/expenses/"+invalidID, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected BadRequest for ID: "+invalidID)

			var jsonResponse map[string]string
			err := json.Unmarshal(rr.Body.Bytes(), &jsonResponse)
			assert.NoError(t, err, "Failed to unmarshal error response for ID: "+invalidID)
			assert.Contains(t, jsonResponse["error"], "Invalid expense ID format", "Expected error message for ID: "+invalidID)
		})
	}
}

func TestUpdateExpenseHandler_InvalidIDFormat(t *testing.T) {
	router, _ := setupExpenseTestRouter(t)

	for _, invalidID := range invalidIDs {
		t.Run(fmt.Sprintf("ID_%s", invalidID), func(t *testing.T) {
			payload := bytes.NewBufferString(`{"note": "test update"}`)
			req, _ := http.NewRequest("PUT", "/expenses/"+invalidID, payload)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected BadRequest for ID: "+invalidID)

			var jsonResponse map[string]string
			err := json.Unmarshal(rr.Body.Bytes(), &jsonResponse)
			assert.NoError(t, err, "Failed to unmarshal error response for ID: "+invalidID)
			assert.Contains(t, jsonResponse["error"], "Invalid expense ID format", "Expected error message for ID: "+invalidID)
		})
	}
}

func TestDeleteExpenseHandler_InvalidIDFormat(t *testing.T) {
	router, _ := setupExpenseTestRouter(t)

	for _, invalidID := range invalidIDs {
		t.Run(fmt.Sprintf("ID_%s", invalidID), func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/expenses/"+invalidID, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected BadRequest for ID: "+invalidID)

			var jsonResponse map[string]string
			err := json.Unmarshal(rr.Body.Bytes(), &jsonResponse)
			assert.NoError(t, err, "Failed to unmarshal error response for ID: "+invalidID)
			assert.Contains(t, jsonResponse["error"], "Invalid expense ID format", "Expected error message for ID: "+invalidID)
		})
	}
}

func TestDeleteExpenseHandler_ValidIDFormat_PassesParsing(t *testing.T) {
	router, db := setupExpenseTestRouter(t)

	// Create a dummy expense to ensure the handler attempts deletion
	// and doesn't fail *before* ID parsing due to a DB error on a non-existent item.
	// However, the service's DeleteExpense itself handles "not found".
	// The goal here is just that ID parsing doesn't yield BadRequest.
	dummyExpense := models.Expense{Amount: 1, Category: "test", Date: database.CustomDate{Time: time.Now()}}
	db.Create(&dummyExpense) // ID will be 1 (or more, depending on test execution order if DB wasn't perfectly isolated before)

	validID := strconv.Itoa(int(dummyExpense.ID)) // Use a real ID from the DB

	req, _ := http.NewRequest("DELETE", "/expenses/"+validID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// We expect StatusNoContent if delete is successful, or StatusNotFound if item was already deleted
	// but crucially NOT StatusBadRequest due to ID parsing.
	assert.NotEqual(t, http.StatusBadRequest, rr.Code, "Expected not BadRequest for valid ID: "+validID)
	if rr.Code == http.StatusBadRequest {
		var jsonResponse map[string]string
		json.Unmarshal(rr.Body.Bytes(), &jsonResponse)
		t.Logf("Unexpected BadRequest response for valid ID %s: %v", validID, jsonResponse["error"])
	}
	// More specific check:
	assert.Condition(t, func() bool {
		return rr.Code == http.StatusNoContent || rr.Code == http.StatusNotFound
	}, "Expected StatusNoContent or StatusNotFound for valid ID %s, but got %d", validID, rr.Code)
}

func TestGetExpenseHandler_ValidIDFormat_PassesParsing(t *testing.T) {
	router, db := setupExpenseTestRouter(t)
	dummyExpense := models.Expense{Amount: 1, Category: "test", Date: database.CustomDate{Time: time.Now()}}
	db.Create(&dummyExpense)
	validID := strconv.Itoa(int(dummyExpense.ID))

	req, _ := http.NewRequest("GET", "/expenses/"+validID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.NotEqual(t, http.StatusBadRequest, rr.Code, "Expected not BadRequest for valid ID: "+validID)
	assert.Condition(t, func() bool {
		return rr.Code == http.StatusOK || rr.Code == http.StatusNotFound // OK if found, NotFound if somehow deleted by another test (unlikely with t.Name in DSN)
	}, "Expected StatusOK or StatusNotFound for valid ID %s, but got %d", validID, rr.Code)
}

func TestUpdateExpenseHandler_ValidIDFormat_PassesParsing(t *testing.T) {
	router, db := setupExpenseTestRouter(t)
	dummyExpense := models.Expense{Amount: 1, Category: "test", Date: database.CustomDate{Time: time.Now()}}
	db.Create(&dummyExpense)
	validID := strconv.Itoa(int(dummyExpense.ID))

	payload := bytes.NewBufferString(`{"note": "test update for valid id"}`)
	req, _ := http.NewRequest("PUT", "/expenses/"+validID, payload)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.NotEqual(t, http.StatusBadRequest, rr.Code, "Expected not BadRequest for valid ID: "+validID)
	assert.Condition(t, func() bool {
		return rr.Code == http.StatusOK || rr.Code == http.StatusNotFound
	}, "Expected StatusOK or StatusNotFound for valid ID %s, but got %d", validID, rr.Code)
}
