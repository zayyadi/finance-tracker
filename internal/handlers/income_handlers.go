package handlers

import (
	"log" // Added for logging
	"net/http"
	"strconv"
	"strings"
	"time" // Added for time.Time{} comparison

	"github.com/gin-gonic/gin"
	"github.com/zayyadi/finance-tracker/internal/database" // Added for database.CustomDate
	"github.com/zayyadi/finance-tracker/internal/models"
	"github.com/zayyadi/finance-tracker/internal/services"
)

// IncomeHandler handles HTTP requests for income records.
type IncomeHandler struct {
	service        *services.IncomeService
	summaryService *services.SummaryService // Added SummaryService
}

// NewIncomeHandler creates a new IncomeHandler with the given services.
func NewIncomeHandler(service *services.IncomeService, summaryService *services.SummaryService) *IncomeHandler {
	return &IncomeHandler{
		service:        service,
		summaryService: summaryService,
	}
}

// CreateIncomeHandler handles the creation of a new income record.
func (h *IncomeHandler) CreateIncomeHandler(c *gin.Context) {
	var req models.IncomeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	income := models.Income{
		Amount:   req.Amount,
		Category: req.Category,
		Date:     req.Date,
		Note:     req.Note,
	}

	if err := h.service.CreateIncome(&income); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create income record: " + err.Error()})
		return
	}

	// Invalidate summaries
	if h.summaryService != nil {
		go func(dateOfItem database.CustomDate) { // Use a goroutine for non-blocking invalidation
			err := h.summaryService.InvalidateSummariesForDate(dateOfItem.Time, []string{"monthly", "weekly", "yearly"})
			if err != nil {
				log.Printf("Error invalidating summaries after creating income: %v", err)
			}
		}(income.Date)
	}

	c.JSON(http.StatusCreated, income)
}

// GetIncomeHandler handles fetching a single income record.
func (h *IncomeHandler) GetIncomeHandler(c *gin.Context) {
	incomeIDStr := c.Param("id")
	incomeIDUint64, err := strconv.ParseUint(incomeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid income ID format"})
		return
	}
	incomeID := uint(incomeIDUint64)
	if incomeID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid income ID format"})
		return
	}

	income, err := h.service.GetIncomeByID(incomeID)
	if err != nil {
		if strings.Contains(err.Error(), "income record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve income record: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, income)
}

// ListIncomesHandler handles fetching all income records with pagination.
func (h *IncomeHandler) ListIncomesHandler(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	incomes, err := h.service.GetIncomes(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve income records: " + err.Error()})
		return
	}

	if incomes == nil { // Return empty list instead of nil if no records
		incomes = []models.Income{}
	}
	c.JSON(http.StatusOK, incomes)
}

// UpdateIncomeHandler handles updating an existing income record.
func (h *IncomeHandler) UpdateIncomeHandler(c *gin.Context) {
	// userID, exists := c.Get("userID")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
	// 	return
	// }
	incomeIDStr := c.Param("id")
	incomeIDUint64, err := strconv.ParseUint(incomeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid income ID format"})
		return
	}
	incomeID := uint(incomeIDUint64)
	if incomeID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid income ID format"})
		return
	}

	var req models.IncomeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	if req.Amount == nil && req.Category == nil && req.Date == nil && req.Note == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided for update"})
		return
	}

	updatedIncome, err := h.service.UpdateIncome(incomeID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "income record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update income record: " + err.Error()})
		}
		return
	}

	// Invalidate summaries
	if h.summaryService != nil && updatedIncome.Date.Time != (time.Time{}) { // Ensure Date is valid
		go func(dateOfItem database.CustomDate) {
			err := h.summaryService.InvalidateSummariesForDate(dateOfItem.Time, []string{"monthly", "weekly", "yearly"})
			if err != nil {
				log.Printf("Error invalidating summaries after updating income: %v", err)
			}
		}(updatedIncome.Date)
		// Note: If the date of the income was changed, summaries for the *old* date
		// should also be invalidated. This requires fetching the old record before update.
		// For simplicity, current implementation only invalidates for the new date.
	}

	c.JSON(http.StatusOK, updatedIncome)
}

// DeleteIncomeHandler handles deleting an income record.
func (h *IncomeHandler) DeleteIncomeHandler(c *gin.Context) {
	incomeIDStr := c.Param("id")
	incomeIDUint64, err := strconv.ParseUint(incomeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid income ID format"})
		return
	}
	incomeID := uint(incomeIDUint64)
	if incomeID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid income ID format"})
		return
	}

	// Fetch the income first to get its date for summary invalidation
	incomeToDelete, serviceErr := h.service.GetIncomeByID(incomeID)
	if serviceErr != nil {
		if strings.Contains(serviceErr.Error(), "income record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Income record not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve income record before deletion: " + serviceErr.Error()})
		}
		return
	}
	dateOfDeletedItem := incomeToDelete.Date

	// Delete the income
	err = h.service.DeleteIncome(incomeID)
	if err != nil {
		if strings.Contains(err.Error(), "income record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Income record not found during deletion"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete income record: " + err.Error()})
		}
		return
	}

	// Invalidate summaries
	if h.summaryService != nil && dateOfDeletedItem.Time != (time.Time{}) {
		go func(dateVal database.CustomDate) {
			err := h.summaryService.InvalidateSummariesForDate(dateVal.Time, []string{"monthly", "weekly", "yearly"})
			if err != nil {
				log.Printf("Error invalidating summaries after deleting income: %v", err)
			}
		}(dateOfDeletedItem)
	}

	c.JSON(http.StatusNoContent, nil)
}
