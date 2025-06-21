package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zayyadi/finance-tracker/internal/services"
)

// AIAdviceHandler handles HTTP requests for AI-based financial advice.
type AIAdviceHandler struct {
	aiAdviceService *services.AIAdviceService
	summaryService  *services.SummaryService
}

// NewAIAdviceHandler creates a new AIAdviceHandler.
func NewAIAdviceHandler(aiService *services.AIAdviceService, sumService *services.SummaryService) *AIAdviceHandler {
	return &AIAdviceHandler{
		aiAdviceService: aiService,
		summaryService:  sumService,
	}
}

// GetAdviceHandler fetches the latest monthly summary and then gets AI advice.
func (h *AIAdviceHandler) GetAdviceHandler(c *gin.Context) {
	// userID, err := GetUserIDFromContext(c) // UserID no longer needed for service call
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found: " + err.Error()})
	// 	return
	// }

	// Fetch the latest monthly summary (e.g., for the current month)
	targetDate := time.Now()
	viewType := "overall" // AI advice should be based on the overall summary
	summary, err := h.summaryService.GetOrCreateFinancialSummary("monthly", targetDate, viewType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get financial summary: " + err.Error()})
		return
	}

	// If summary was fetched/created successfully, get advice
	advice, err := h.aiAdviceService.GetFinancialAdvice(summary)
	if err != nil {
		// Check if it's the "API key not set" specific message to return a more user-friendly error
		if err.Error() == "OPENROUTER_API_KEY is not set" || advice == "AI features are currently unavailable as the API key is not configured." {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI advice feature is not configured."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get financial advice: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"advice": advice})
}
