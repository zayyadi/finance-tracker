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

// setupIncomeTestRouter initializes an in-memory SQLite database and sets up the Gin router
// with income routes for testing.
func setupIncomeTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:inc_handler_%s_%d?mode=memory&cache=shared", t.Name(), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err, "Failed to connect to in-memory SQLite")

	sqlDB, errDB := db.DB()
	assert.NoError(t, errDB)
	t.Cleanup(func() { sqlDB.Close() })

	db.Exec("DROP TABLE IF EXISTS Income")
	db.Exec("DROP TABLE IF EXISTS Users")

	err = db.AutoMigrate(&models.User{}, &models.Income{})
	assert.NoError(t, err, "Failed to auto-migrate models")

	db.Create(&models.User{Username: "testuser", Email: "test@example.com", PasswordHash: "hash"})

	incomeService := services.NewIncomeService(db)
	incomeHandler := NewIncomeHandler(incomeService)

	router := gin.Default()
	router.GET("/income/:id", incomeHandler.GetIncomeHandler)
	router.PUT("/income/:id", incomeHandler.UpdateIncomeHandler)
	router.DELETE("/income/:id", incomeHandler.DeleteIncomeHandler)
	// Add other routes like POST /income, GET /income if needed by specific valid ID tests for setup.

	return router, db
}

var testInvalidIDs = []string{"not-a-number", " ", "1.0", "1a2b", "-1", "0"}

func TestGetIncomeHandler_InvalidIDFormat(t *testing.T) {
	router, _ := setupIncomeTestRouter(t)

	for _, invalidID := range testInvalidIDs {
		t.Run(fmt.Sprintf("ID_%s", invalidID), func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/income/"+invalidID, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected BadRequest for ID: "+invalidID)

			var jsonResponse map[string]string
			err := json.Unmarshal(rr.Body.Bytes(), &jsonResponse)
			assert.NoError(t, err, "Failed to unmarshal error response for ID: "+invalidID)
			assert.Contains(t, jsonResponse["error"], "Invalid income ID format", "Expected error message for ID: "+invalidID)
		})
	}
}

func TestUpdateIncomeHandler_InvalidIDFormat(t *testing.T) {
	router, _ := setupIncomeTestRouter(t)

	for _, invalidID := range testInvalidIDs {
		t.Run(fmt.Sprintf("ID_%s", invalidID), func(t *testing.T) {
			payload := bytes.NewBufferString(`{"note": "test update"}`)
			req, _ := http.NewRequest("PUT", "/income/"+invalidID, payload)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected BadRequest for ID: "+invalidID)

			var jsonResponse map[string]string
			err := json.Unmarshal(rr.Body.Bytes(), &jsonResponse)
			assert.NoError(t, err, "Failed to unmarshal error response for ID: "+invalidID)
			assert.Contains(t, jsonResponse["error"], "Invalid income ID format", "Expected error message for ID: "+invalidID)
		})
	}
}

func TestDeleteIncomeHandler_InvalidIDFormat(t *testing.T) {
	router, _ := setupIncomeTestRouter(t)

	for _, invalidID := range testInvalidIDs {
		t.Run(fmt.Sprintf("ID_%s", invalidID), func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/income/"+invalidID, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected BadRequest for ID: "+invalidID)

			var jsonResponse map[string]string
			err := json.Unmarshal(rr.Body.Bytes(), &jsonResponse)
			assert.NoError(t, err, "Failed to unmarshal error response for ID: "+invalidID)
			assert.Contains(t, jsonResponse["error"], "Invalid income ID format", "Expected error message for ID: "+invalidID)
		})
	}
}

func TestGetIncomeHandler_ValidIDFormat_PassesParsing(t *testing.T) {
	router, db := setupIncomeTestRouter(t)
	dummyIncome := models.Income{Amount: 1, Category: "test", Date: database.CustomDate{Time: time.Now()}}
	db.Create(&dummyIncome)
	validID := strconv.Itoa(int(dummyIncome.ID))

	req, _ := http.NewRequest("GET", "/income/"+validID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.NotEqual(t, http.StatusBadRequest, rr.Code, "Expected not BadRequest for valid ID: "+validID)
	assert.Condition(t, func() bool {
		return rr.Code == http.StatusOK || rr.Code == http.StatusNotFound
	}, "Expected StatusOK or StatusNotFound for valid ID %s, but got %d", validID, rr.Code)
}

func TestUpdateIncomeHandler_ValidIDFormat_PassesParsing(t *testing.T) {
	router, db := setupIncomeTestRouter(t)
	dummyIncome := models.Income{Amount: 1, Category: "test", Date: database.CustomDate{Time: time.Now()}}
	db.Create(&dummyIncome)
	validID := strconv.Itoa(int(dummyIncome.ID))

	payload := bytes.NewBufferString(`{"note": "test update for valid id"}`)
	req, _ := http.NewRequest("PUT", "/income/"+validID, payload)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.NotEqual(t, http.StatusBadRequest, rr.Code, "Expected not BadRequest for valid ID: "+validID)
	assert.Condition(t, func() bool {
		return rr.Code == http.StatusOK || rr.Code == http.StatusNotFound
	}, "Expected StatusOK or StatusNotFound for valid ID %s, but got %d", validID, rr.Code)
}

func TestDeleteIncomeHandler_ValidIDFormat_PassesParsing(t *testing.T) {
	router, db := setupIncomeTestRouter(t)
	dummyIncome := models.Income{Amount: 1, Category: "test", Date: database.CustomDate{Time: time.Now()}}
	db.Create(&dummyIncome)
	validID := strconv.Itoa(int(dummyIncome.ID))

	req, _ := http.NewRequest("DELETE", "/income/"+validID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.NotEqual(t, http.StatusBadRequest, rr.Code, "Expected not BadRequest for valid ID: "+validID)
	assert.Condition(t, func() bool {
		return rr.Code == http.StatusNoContent || rr.Code == http.StatusNotFound
	}, "Expected StatusNoContent or StatusNotFound for valid ID %s, but got %d", validID, rr.Code)
}
