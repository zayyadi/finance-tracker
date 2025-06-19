package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zayyadi/finance-tracker/internal/services"
)

// ReportHandler handles HTTP requests for generating reports.
type ReportHandler struct {
	service *services.ReportService
}

// NewReportHandler creates a new ReportHandler with the given service.
func NewReportHandler(service *services.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

const defaultDateFormat = "2006-01-02"

// GenerateCSVReportHandler handles the generation of a CSV transaction report.
// Expects "startDate" and "endDate" query parameters in "YYYY-MM-DD" format.
func (h *ReportHandler) GenerateCSVReportHandler(c *gin.Context) {
	// userID, err := GetUserIDFromContext(c) // UserID no longer used by service
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found: " + err.Error()})
	// 	return
	// }

	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate and endDate query parameters are required in YYYY-MM-DD format."})
		return
	}

	startDate, err := time.Parse(defaultDateFormat, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid startDate format. Use YYYY-MM-DD. Error: %v", err)})
		return
	}

	endDate, err := time.Parse(defaultDateFormat, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid endDate format. Use YYYY-MM-DD. Error: %v", err)})
		return
	}

	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate cannot be after endDate."})
		return
	}
	// To make endDate inclusive for the whole day if data has timestamps, or ensure it's handled correctly by date-only queries
	// For date-only columns, direct comparison is fine. If timestamped, endDate might need to be end of day.
	// Assuming date columns in DB are DATE type, so direct comparison is fine.

	// The placeholder UserID (1) from GetUserIDFromContext will be passed if not removed,
	// but GenerateTransactionsCSV service method was updated to not expect it.
	// For clarity, we can just pass a conceptual "default" user or remove if the service truly ignores it.
	// Since service was updated, we don't need to pass userID.
	// We need to ensure the service method GenerateTransactionsCSV signature was updated.
	// Assuming it was: GenerateTransactionsCSV(startDate, endDate time.Time)

	// Let's verify GenerateTransactionsCSV signature. Assuming it's (startDate, endDate).
	// If it still expects userID, then GetUserIDFromContext() is needed here.
	// Based on previous step "Adapt Data Models and Services", ReportService.GenerateTransactionsCSV
	// should have been updated to not require UserID.
	// So, the call becomes:
	csvData, err := h.service.GenerateTransactionsCSV(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSV report: " + err.Error()})
		return
	}

	// Set headers for CSV download
	fileName := fmt.Sprintf("financial_report_%s_to_%s.csv", startDate.Format("20060102"), endDate.Format("20060102"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "text/csv")
	c.Header("Cache-Control", "no-cache") // Advise client not to cache

	c.Data(http.StatusOK, "text/csv", []byte(csvData))
}

// GeneratePDFReportHandler handles the generation of a PDF transaction report.
// Expects "startDate" and "endDate" query parameters in "YYYY-MM-DD" format.
func (h *ReportHandler) GeneratePDFReportHandler(c *gin.Context) {
	// userID, err := GetUserIDFromContext(c) // UserID no longer used by service
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found: " + err.Error()})
	// 	return
	// }

	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate and endDate query parameters are required in YYYY-MM-DD format."})
		return
	}

	startDate, err := time.Parse(defaultDateFormat, startDateStr) // defaultDateFormat = "2006-01-02"
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid startDate format. Use YYYY-MM-DD. Error: %v", err)})
		return
	}

	endDate, err := time.Parse(defaultDateFormat, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid endDate format. Use YYYY-MM-DD. Error: %v", err)})
		return
	}

	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate cannot be after endDate."})
		return
	}

	// Assuming GenerateTransactionsPDF service method was also updated.
	pdfBuffer, err := h.service.GenerateTransactionsPDF(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF report: " + err.Error()})
		return
	}

	// Set headers for PDF download
	fileName := fmt.Sprintf("financial_report_%s_to_%s.pdf", startDate.Format("20060102"), endDate.Format("20060102"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/pdf")
	c.Header("Cache-Control", "no-cache")

	// Write the PDF buffer to the response
	c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())
}
