package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zayyadi/finance-tracker/internal/models"
	"github.com/zayyadi/finance-tracker/internal/services"
)

// DebtHandler handles HTTP requests for debt records.
type DebtHandler struct {
	service *services.DebtService
}

// NewDebtHandler creates a new DebtHandler with the given service.
func NewDebtHandler(service *services.DebtService) *DebtHandler {
	return &DebtHandler{service: service}
}

// CreateDebtHandler handles the creation of a new debt record.
func (h *DebtHandler) CreateDebtHandler(c *gin.Context) {
	// _, err := GetUserIDFromContext(c) // UserID no longer needed
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	// 	return
	// }

	var req models.DebtCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}
	status := "Pending" // Default if not provided or if GORM default doesn't kick in via struct
	if req.Status != nil && *req.Status != "" {
		status = *req.Status
	}

	debt := models.Debt{
		// UserID:      userID, // UserID removed
		DebtorName:  req.DebtorName,
		Description: description,
		Amount:      req.Amount,
		DueDate:     req.DueDate,
		Status:      status,
	}

	if err := h.service.CreateDebt(&debt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create debt record: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, debt)
}

// GetDebtHandler handles fetching a single debt record.
func (h *DebtHandler) GetDebtHandler(c *gin.Context) {
	// _, err := GetUserIDFromContext(c) // UserID no longer needed
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	// 	return
	// }

	debtIDStr := c.Param("id")
	debtIDUint64, err := strconv.ParseUint(debtIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid debt ID format"})
		return
	}
	debtID := uint(debtIDUint64)
	if debtID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid debt ID format"})
		return
	}

	debt, err := h.service.GetDebtByID(debtID) // UserID removed
	if err != nil {
		if strings.Contains(err.Error(), "debt record not found") { // Updated error check
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve debt record: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, debt)
}

// ListDebtsHandler handles fetching all debt records with pagination and optional status filter.
func (h *DebtHandler) ListDebtsHandler(c *gin.Context) {
	// _, err := GetUserIDFromContext(c) // UserID no longer needed
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	// 	return
	// }

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	statusFilter := strings.TrimSpace(c.Query("status"))

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	if statusFilter != "" && !(statusFilter == "Pending" || statusFilter == "Paid" || statusFilter == "Overdue") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status filter. Allowed values: Pending, Paid, Overdue"})
		return
	}

	debts, err := h.service.GetDebts(offset, limit, statusFilter) // Changed from GetDebtsByUser
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve debt records: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, debts)
}

// UpdateDebtHandler handles updating an existing debt record.
func (h *DebtHandler) UpdateDebtHandler(c *gin.Context) {
	// _, err := GetUserIDFromContext(c) // UserID no longer needed
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	// 	return
	// }

	debtIDStr := c.Param("id")
	debtIDUint64, err := strconv.ParseUint(debtIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid debt ID format"})
		return
	}
	debtID := uint(debtIDUint64)
	if debtID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid debt ID format"})
		return
	}

	var req models.DebtUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	if req.DebtorName == nil && req.Description == nil && req.Amount == nil &&
		req.DueDate == nil && req.Status == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided for update"})
		return
	}

	updatedDebt, err := h.service.UpdateDebt(debtID, &req) // UserID removed
	if err != nil {
		if strings.Contains(err.Error(), "debt record not found") { // Updated error check
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update debt record: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updatedDebt)
}

// DeleteDebtHandler handles deleting a debt record.
func (h *DebtHandler) DeleteDebtHandler(c *gin.Context) {
	// _, err := GetUserIDFromContext(c) // UserID no longer needed
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	// 	return
	// }

	debtIDStr := c.Param("id")
	debtIDUint64, err := strconv.ParseUint(debtIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid debt ID format"})
		return
	}
	debtID := uint(debtIDUint64)
	if debtID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid debt ID format"})
		return
	}

	err = h.service.DeleteDebt(debtID) // UserID removed
	if err != nil {
		if strings.Contains(err.Error(), "debt record not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Debt record not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete debt record: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
