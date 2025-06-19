package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zayyadi/finance-tracker/internal/models"

	// "fmt" // No longer needed here if getUserIDFromContext is in this package
	// "github.com/zayyadi/finance-tracker/internal/models"
	"github.com/zayyadi/finance-tracker/internal/services"
)

// IncomeHandler handles HTTP requests for income records.
type IncomeHandler struct {
	service *services.IncomeService
}

// NewIncomeHandler creates a new IncomeHandler with the given service.
func NewIncomeHandler(service *services.IncomeService) *IncomeHandler {
	return &IncomeHandler{service: service}
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

	err = h.service.DeleteIncome(incomeID)
	if err != nil {
		if strings.Contains(err.Error(), "income record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Income record not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete income record: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
