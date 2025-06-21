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

// setupDebtTestRouter initializes an in-memory SQLite database and sets up the Gin router
// with debt routes for testing.
func setupDebtTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:debt_handler_%s_%d?mode=memory&cache=shared", t.Name(), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err, "Failed to connect to in-memory SQLite")

	sqlDB, errDB := db.DB()
	assert.NoError(t, errDB)
	t.Cleanup(func() { sqlDB.Close() })

	db.Exec("DROP TABLE IF EXISTS Debts")
	db.Exec("DROP TABLE IF EXISTS Users")

	err = db.AutoMigrate(&models.User{}, &models.Debt{})
	assert.NoError(t, err, "Failed to auto-migrate models")

	db.Create(&models.User{Username: "testuser", Email: "test@example.com", PasswordHash: "hash"})

	debtService := services.NewDebtService(db)
	debtHandler := NewDebtHandler(debtService)

	router := gin.Default()
	router.GET("/debts/:id", debtHandler.GetDebtHandler)
	router.PUT("/debts/:id", debtHandler.UpdateDebtHandler)
	router.DELETE("/debts/:id", debtHandler.DeleteDebtHandler)
	// Add other routes like POST /debts, GET /debts if needed by specific valid ID tests for setup.


	return router, db
}

var testInvalidDebtIDs = []string{"not-a-number", " ", "1.0", "1a2b", "-1", "0"}

func TestGetDebtHandler_InvalidIDFormat(t *testing.T) {
	router, _ := setupDebtTestRouter(t)

	for _, invalidID := range testInvalidDebtIDs {
		t.Run(fmt.Sprintf("ID_%s", invalidID), func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/debts/"+invalidID, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected BadRequest for ID: "+invalidID)

			var jsonResponse map[string]string
			err := json.Unmarshal(rr.Body.Bytes(), &jsonResponse)
			assert.NoError(t, err, "Failed to unmarshal error response for ID: "+invalidID)
			assert.Contains(t, jsonResponse["error"], "Invalid debt ID format", "Expected error message for ID: "+invalidID)
		})
	}
}

func TestUpdateDebtHandler_InvalidIDFormat(t *testing.T) {
	router, _ := setupDebtTestRouter(t)

	for _, invalidID := range testInvalidDebtIDs {
		t.Run(fmt.Sprintf("ID_%s", invalidID), func(t *testing.T) {
			payload := bytes.NewBufferString(`{"description": "test update"}`)
			req, _ := http.NewRequest("PUT", "/debts/"+invalidID, payload)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected BadRequest for ID: "+invalidID)

			var jsonResponse map[string]string
			err := json.Unmarshal(rr.Body.Bytes(), &jsonResponse)
			assert.NoError(t, err, "Failed to unmarshal error response for ID: "+invalidID)
			assert.Contains(t, jsonResponse["error"], "Invalid debt ID format", "Expected error message for ID: "+invalidID)
		})
	}
}

func TestDeleteDebtHandler_InvalidIDFormat(t *testing.T) {
	router, _ := setupDebtTestRouter(t)

	for _, invalidID := range testInvalidDebtIDs {
		t.Run(fmt.Sprintf("ID_%s", invalidID), func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/debts/"+invalidID, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected BadRequest for ID: "+invalidID)

			var jsonResponse map[string]string
			err := json.Unmarshal(rr.Body.Bytes(), &jsonResponse)
			assert.NoError(t, err, "Failed to unmarshal error response for ID: "+invalidID)
			assert.Contains(t, jsonResponse["error"], "Invalid debt ID format", "Expected error message for ID: "+invalidID)
		})
	}
}

func TestGetDebtHandler_ValidIDFormat_PassesParsing(t *testing.T) {
	router, db := setupDebtTestRouter(t)
	// DueDate requires a valid date, Amount > 0
	validDueDate, _ := time.Parse("2006-01-02", "2025-01-01")
	dummyDebt := models.Debt{DebtorName:"Test Debtor", Amount: 100, DueDate: database.CustomDate{Time: validDueDate}, Status: "Pending"}
	db.Create(&dummyDebt)
	validID := strconv.Itoa(int(dummyDebt.ID))

	req, _ := http.NewRequest("GET", "/debts/"+validID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.NotEqual(t, http.StatusBadRequest, rr.Code, "Expected not BadRequest for valid ID: "+validID)
	assert.Condition(t, func() bool {
		return rr.Code == http.StatusOK || rr.Code == http.StatusNotFound
	}, "Expected StatusOK or StatusNotFound for valid ID %s, but got %d", validID, rr.Code)
}

func TestUpdateDebtHandler_ValidIDFormat_PassesParsing(t *testing.T) {
	router, db := setupDebtTestRouter(t)
	validDueDate, _ := time.Parse("2006-01-02", "2025-01-01")
	dummyDebt := models.Debt{DebtorName:"Test Debtor", Amount: 100, DueDate: database.CustomDate{Time: validDueDate}, Status: "Pending"}
	db.Create(&dummyDebt)
	validID := strconv.Itoa(int(dummyDebt.ID))

	payload := bytes.NewBufferString(`{"description": "test update for valid id"}`)
	req, _ := http.NewRequest("PUT", "/debts/"+validID, payload)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.NotEqual(t, http.StatusBadRequest, rr.Code, "Expected not BadRequest for valid ID: "+validID)
	assert.Condition(t, func() bool {
		return rr.Code == http.StatusOK || rr.Code == http.StatusNotFound
	}, "Expected StatusOK or StatusNotFound for valid ID %s, but got %d", validID, rr.Code)
}

func TestDeleteDebtHandler_ValidIDFormat_PassesParsing(t *testing.T) {
	router, db := setupDebtTestRouter(t)
	validDueDate, _ := time.Parse("2006-01-02", "2025-01-01")
	dummyDebt := models.Debt{DebtorName:"Test Debtor", Amount: 100, DueDate: database.CustomDate{Time: validDueDate}, Status: "Pending"}
	db.Create(&dummyDebt)
	validID := strconv.Itoa(int(dummyDebt.ID))

	req, _ := http.NewRequest("DELETE", "/debts/"+validID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.NotEqual(t, http.StatusBadRequest, rr.Code, "Expected not BadRequest for valid ID: "+validID)
	assert.Condition(t, func() bool {
		return rr.Code == http.StatusNoContent || rr.Code == http.StatusNotFound
	}, "Expected StatusNoContent or StatusNotFound for valid ID %s, but got %d", validID, rr.Code)
}
