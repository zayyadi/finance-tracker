package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zayyadi/finance-tracker/internal/services"
)

// AnalyticsHandler handles analytics-related API requests.
type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(service *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: service}
}

// GetExpenseBreakdownHandler handles requests for expense breakdown by category.
// @Summary Get expense breakdown by category for the current month
// @Description Retrieves total expenses for each category for the current calendar month.
// @Tags analytics
// @Produce json
// @Success 200 {array} models.CategoryExpenseStat
// @Failure 500 {object} ErrorResponse
// @Router /analytics/expense-categories [get]
func (h *AnalyticsHandler) GetExpenseBreakdownHandler(c *gin.Context) {
	// For now, use the current date to determine the target month.
	// This could be extended to accept a date query parameter.
	targetDate := time.Now()

	stats, err := h.analyticsService.GetExpenseBreakdownByCategory(targetDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get expense breakdown: " + err.Error()})
		return
	}

	if stats == nil {
		// Return empty array instead of null if no stats
		c.JSON(http.StatusOK, []gin.H{})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetIncomeExpenseTrendHandler handles requests for income vs. expense trends.
// @Summary Get income vs. expense trend for a number of past months
// @Description Retrieves total income and expenses for a specified number of past months, including the current month.
// @Tags analytics
// @Produce json
// @Param months query int false "Number of months for the trend (default: 6)"
// @Success 200 {array} models.MonthlyTrendStat
// @Failure 400 {object} ErrorResponse "Invalid number of months"
// @Failure 500 {object} ErrorResponse
// @Router /analytics/income-expense-trend [get]
func (h *AnalyticsHandler) GetIncomeExpenseTrendHandler(c *gin.Context) {
	numMonthsStr := c.DefaultQuery("months", "6") // Default to 6 months
	numMonths, err := strconv.Atoi(numMonthsStr)
	if err != nil || numMonths <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid number of months specified. Must be a positive integer."})
		return
	}

	trend, err := h.analyticsService.GetIncomeExpenseTrend(numMonths)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get income-expense trend: " + err.Error()})
		return
	}

	if trend == nil {
		// Return empty array instead of null if no trend data
		c.JSON(http.StatusOK, []gin.H{})
		return
	}
	c.JSON(http.StatusOK, trend)
}

// ErrorResponse is a generic structure for error responses.
// type ErrorResponse struct {
//	 Error string `json:"error"`
// }
// This can be uncommented and used if a more structured error is needed. Helper functions can also be created.
