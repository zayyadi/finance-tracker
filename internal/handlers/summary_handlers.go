package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zayyadi/finance-tracker/internal/models"
	"github.com/zayyadi/finance-tracker/internal/services"
)

// SummaryHandler handles HTTP requests for financial summaries.
type SummaryHandler struct {
	service *services.SummaryService
}

// NewSummaryHandler creates a new SummaryHandler with the given service.
func NewSummaryHandler(service *services.SummaryService) *SummaryHandler {
	return &SummaryHandler{service: service}
}

// GetMonthlySummaryHandler handles requests for monthly financial summaries.
// Expects a "date" query parameter in "YYYY-MM" format.
func (h *SummaryHandler) GetMonthlySummaryHandler(c *gin.Context) {
	// _, err := GetUserIDFromContext(c) // UserID no longer needed for service call
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	// 	return
	// }

	var req models.SummaryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date query parameter is required: " + err.Error()})
		return
	}

	targetDate, err := time.Parse("2006-01", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format for monthly summary. Use YYYY-MM."})
		return
	}

	view := c.DefaultQuery("view", "overall")
	allowedViews := map[string]bool{"overall": true, "income": true, "expenses": true, "savings": true, "debts": true}
	if !allowedViews[view] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid view type specified"})
		return
	}

	summary, err := h.service.GetOrCreateFinancialSummary("monthly", targetDate, view)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get or create monthly summary: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// GetWeeklySummaryHandler handles requests for weekly financial summaries.
// Expects a "date" query parameter in "YYYY-MM-DD" format (any date within the desired week).
func (h *SummaryHandler) GetWeeklySummaryHandler(c *gin.Context) {
	// _, err := GetUserIDFromContext(c) // UserID no longer needed
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	// 	return
	// }

	var req models.SummaryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date query parameter is required: " + err.Error()})
		return
	}

	targetDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format for weekly summary. Use YYYY-MM-DD."})
		return
	}

	view := c.DefaultQuery("view", "overall")
	allowedViews := map[string]bool{"overall": true, "income": true, "expenses": true, "savings": true, "debts": true}
	if !allowedViews[view] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid view type specified"})
		return
	}

	summary, err := h.service.GetOrCreateFinancialSummary("weekly", targetDate, view)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get or create weekly summary: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// GetYearlySummaryHandler handles requests for yearly financial summaries.
// Expects a "date" query parameter in "YYYY" format.
func (h *SummaryHandler) GetYearlySummaryHandler(c *gin.Context) {
	// _, err := GetUserIDFromContext(c) // UserID no longer needed
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	// 	return
	// }

	var req models.SummaryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date query parameter is required: " + err.Error()})
		return
	}

	targetDate, err := time.Parse("2006", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format for yearly summary. Use YYYY."})
		return
	}

	view := c.DefaultQuery("view", "overall")
	allowedViews := map[string]bool{"overall": true, "income": true, "expenses": true, "savings": true, "debts": true}
	if !allowedViews[view] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid view type specified"})
		return
	}

	summary, err := h.service.GetOrCreateFinancialSummary("yearly", targetDate, view)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get or create yearly summary: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}
