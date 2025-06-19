package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zayyadi/finance-tracker/internal/models"
	"github.com/zayyadi/finance-tracker/internal/services"
)

// ExpenseHandler handles HTTP requests for expense records.
type ExpenseHandler struct {
	service *services.ExpenseService
}

// NewExpenseHandler creates a new ExpenseHandler with the given service.
func NewExpenseHandler(service *services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: service}
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

	expenses, err := h.service.GetExpenses(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expense records: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, expenses)
}

// UpdateExpenseHandler handles updating an existing expense record.
func (h *ExpenseHandler) UpdateExpenseHandler(c *gin.Context) {
	expenseIDStr := c.Param("id")
	expenseIDUint64, err := strconv.ParseUint(expenseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID format"})
		return
	}
	expenseID := uint(expenseIDUint64)

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

	c.JSON(http.StatusOK, updatedExpense)
}

// DeleteExpenseHandler handles deleting an expense record.
func (h *ExpenseHandler) DeleteExpenseHandler(c *gin.Context) {
	expenseIDStr := c.Param("id")
	expenseIDUint64, err := strconv.ParseUint(expenseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID format"})
		return
	}
	expenseID := uint(expenseIDUint64)

	err = h.service.DeleteExpense(expenseID)
	if err != nil {
		if strings.Contains(err.Error(), "expense record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Expense record not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete expense record: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
