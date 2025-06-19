package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zayyadi/finance-tracker/internal/models"
	"github.com/zayyadi/finance-tracker/internal/services"
)

// SavingsHandler handles HTTP requests for savings goal records.
type SavingsHandler struct {
	service *services.SavingsService
}

// NewSavingsHandler creates a new SavingsHandler with the given service.
func NewSavingsHandler(service *services.SavingsService) *SavingsHandler {
	return &SavingsHandler{service: service}
}

// CreateSavingsHandler handles the creation of a new savings goal.
func (h *SavingsHandler) CreateSavingsHandler(c *gin.Context) {
	var req models.SavingsCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	currentAmount := 0.0
	if req.CurrentAmount != nil {
		currentAmount = *req.CurrentAmount
	}
	notes := ""
	if req.Notes != nil {
		notes = *req.Notes
	}

	savings := models.Savings{
		GoalName:      req.GoalName,
		GoalAmount:    req.GoalAmount,
		CurrentAmount: currentAmount,
		StartDate:     req.StartDate,
		TargetDate:    req.TargetDate,
		Notes:         notes,
	}

	if err := h.service.CreateSavings(&savings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create savings goal: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, savings)
}

// GetSavingsHandler handles fetching a single savings goal.
func (h *SavingsHandler) GetSavingsHandler(c *gin.Context) {
	savingsIDStr := c.Param("id")
	savingsIDUint64, err := strconv.ParseUint(savingsIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid savings ID format"})
		return
	}
	savingsID := uint(savingsIDUint64)

	savings, err := h.service.GetSavingsByID(savingsID)
	if err != nil {
		if strings.Contains(err.Error(), "savings goal not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve savings goal: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, savings)
}

// ListSavingsHandler handles fetching all savings goals with pagination.
func (h *SavingsHandler) ListSavingsHandler(c *gin.Context) {
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

	savingsList, err := h.service.GetSavings(offset, limit) // Changed from GetSavingsByUser
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve savings goals: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, savingsList)
}

// UpdateSavingsHandler handles updating an existing savings goal.
func (h *SavingsHandler) UpdateSavingsHandler(c *gin.Context) {
	savingsIDStr := c.Param("id")
	savingsIDUint64, err := strconv.ParseUint(savingsIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid savings ID format"})
		return
	}
	savingsID := uint(savingsIDUint64)

	var req models.SavingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	if req.GoalName == nil && req.GoalAmount == nil && req.CurrentAmount == nil &&
		req.StartDate == nil && req.TargetDate == nil && req.Notes == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided for update"})
		return
	}

	updatedSavings, err := h.service.UpdateSavings(savingsID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "savings goal not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update savings goal: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updatedSavings)
}

// DeleteSavingsHandler handles deleting a savings goal.
func (h *SavingsHandler) DeleteSavingsHandler(c *gin.Context) {
	savingsIDStr := c.Param("id")
	savingsIDUint64, err := strconv.ParseUint(savingsIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid savings ID format"})
		return
	}
	savingsID := uint(savingsIDUint64)

	err = h.service.DeleteSavings(savingsID)
	if err != nil {
		if strings.Contains(err.Error(), "savings goal not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Savings goal not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete savings goal: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
