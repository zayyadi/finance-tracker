package handlers

import (
	"log" // Added for logging
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zayyadi/finance-tracker/internal/models"
	"github.com/zayyadi/finance-tracker/internal/services"
)

// ExpenseHandler handles HTTP requests for expense records.
type ExpenseHandler struct {
	service        *services.ExpenseService
	summaryService *services.SummaryService // Added SummaryService
}

// NewExpenseHandler creates a new ExpenseHandler with the given services.
func NewExpenseHandler(service *services.ExpenseService, summaryService *services.SummaryService) *ExpenseHandler {
	return &ExpenseHandler{
		service:        service,
		summaryService: summaryService,
	}
}

// CreateExpenseHandler handles the creation of a new expense record.
func (h *ExpenseHandler) CreateExpenseHandler(c *gin.Context) {
	var req models.ExpenseCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	expense := models.Expense{
		Amount:   req.Amount,
		Category: req.Category,
		Date:     req.Date,
		Note:     req.Note,
	}

	if err := h.service.CreateExpense(&expense); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expense record: " + err.Error()})
		return
	}

	// Invalidate summaries
	if h.summaryService != nil {
		go func(dateOfItem database.CustomDate) { // Use a goroutine for non-blocking invalidation
			err := h.summaryService.InvalidateSummariesForDate(dateOfItem.Time, []string{"monthly", "weekly", "yearly"})
			if err != nil {
				log.Printf("Error invalidating summaries after creating expense: %v", err)
			}
		}(expense.Date)
	}

	c.JSON(http.StatusCreated, expense)
}

// GetExpenseHandler handles fetching a single expense record.
func (h *ExpenseHandler) GetExpenseHandler(c *gin.Context) {
	expenseIDStr := c.Param("id")
	expenseIDUint64, err := strconv.ParseUint(expenseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID format"})
		return
	}
	expenseID := uint(expenseIDUint64)
	if expenseID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID format"})
		return
	}

	expense, err := h.service.GetExpenseByID(expenseID)
	if err != nil {
		if strings.Contains(err.Error(), "expense record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expense record: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, expense)
}

// ListExpensesHandler handles fetching all expense records with pagination.
func (h *ExpenseHandler) ListExpensesHandler(c *gin.Context) {
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

	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	expenses, err := h.service.GetExpenses(offset, limit, startDateStr, endDateStr)
	if err != nil {
		// Check if the error is due to date parsing to return a more specific client error
		if strings.Contains(err.Error(), "invalid start date format") || strings.Contains(err.Error(), "invalid end date format") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expense records: " + err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, expenses)
}

// UpdateExpenseHandler handles updating an existing expense record.
func (h *ExpenseHandler) UpdateExpenseHandler(c *gin.Context) {
	expenseIDStr := c.Param("id")
	log.Printf("[ExpenseHandler] UpdateExpenseHandler: Received raw ID string: '%s'", expenseIDStr)
	expenseIDUint64, err := strconv.ParseUint(expenseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID format"})
		return
	}
	expenseID := uint(expenseIDUint64)
	if expenseID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID format"})
		return
	}

	var req models.ExpenseUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	if req.Amount == nil && req.Category == nil && req.Date == nil && req.Note == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided for update"})
		return
	}

	updatedExpense, err := h.service.UpdateExpense(expenseID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "expense record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expense record: " + err.Error()})
		}
		return
	}

	// Invalidate summaries
	if h.summaryService != nil && updatedExpense.Date.Time != (time.Time{}) { // Ensure Date is valid
		go func(dateOfItem database.CustomDate) {
			err := h.summaryService.InvalidateSummariesForDate(dateOfItem.Time, []string{"monthly", "weekly", "yearly"})
			if err != nil {
				log.Printf("Error invalidating summaries after updating expense: %v", err)
			}
		}(updatedExpense.Date)
		// Note: If the date of the expense was changed, summaries for the *old* date
		// should also be invalidated. This requires fetching the old record before update.
		// For simplicity, current implementation only invalidates for the new date.
	}

	c.JSON(http.StatusOK, updatedExpense)
}

// DeleteExpenseHandler handles deleting an expense record.
func (h *ExpenseHandler) DeleteExpenseHandler(c *gin.Context) {
	expenseIDStr := c.Param("id")
	log.Printf("[ExpenseHandler] DeleteExpenseHandler: Received raw ID string: '%s'", expenseIDStr)
	expenseIDUint64, err := strconv.ParseUint(expenseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID format"})
		return
	}
	expenseID := uint(expenseIDUint64)
	if expenseID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID format"})
		return
	}

	// Fetch the expense first to get its date for summary invalidation
	expenseToDelete, serviceErr := h.service.GetExpenseByID(expenseID)
	if serviceErr != nil {
		if strings.Contains(serviceErr.Error(), "expense record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Expense record not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expense record before deletion: " + serviceErr.Error()})
		}
		return
	}
	dateOfDeletedItem := expenseToDelete.Date

	// Delete the expense
	err = h.service.DeleteExpense(expenseID)
	if err != nil {
		// This check might be redundant if GetExpenseByID already confirmed existence,
		// but kept for safety or if DeleteExpense has other failure modes.
		if strings.Contains(err.Error(), "expense record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Expense record not found during deletion"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete expense record: " + err.Error()})
		}
		return
	}

	// Invalidate summaries
	if h.summaryService != nil && dateOfDeletedItem.Time != (time.Time{}) {
		go func(dateVal database.CustomDate) {
			err := h.summaryService.InvalidateSummariesForDate(dateVal.Time, []string{"monthly", "weekly", "yearly"})
			if err != nil {
				log.Printf("Error invalidating summaries after deleting expense: %v", err)
			}
		}(dateOfDeletedItem)
	}

	c.JSON(http.StatusNoContent, nil)
}
