package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"time"

	// For strings.Builder if used instead of bytes.Buffer for CSV
	"github.com/jung-kurt/gofpdf/v2"
)

// ReportService provides methods for generating financial reports.
type ReportService struct {
	incomeService  *IncomeService
	expenseService *ExpenseService
}

// NewReportService creates a new ReportService with necessary service dependencies.
func NewReportService(is *IncomeService, es *ExpenseService) *ReportService {
	return &ReportService{
		incomeService:  is,
		expenseService: es,
	}
}

// GenerateTransactionsCSV generates a CSV string of all income and expense transactions within a date range.
func (s *ReportService) GenerateTransactionsCSV(startDate, endDate time.Time) (string, error) {
	if s.incomeService == nil || s.expenseService == nil {
		return "", fmt.Errorf("report service is not properly initialized with income/expense services")
	}

	incomes, err := s.incomeService.GetIncomesByDateRange(startDate, endDate) // UserID removed
	if err != nil {
		return "", fmt.Errorf("error fetching income data: %w", err)
	}

	expenses, err := s.expenseService.GetExpensesByDateRange(startDate, endDate) // UserID removed
	if err != nil {
		return "", fmt.Errorf("error fetching expense data: %w", err)
	}

	var b bytes.Buffer
	writer := csv.NewWriter(&b)
	defer writer.Flush()

	// Write header row
	header := []string{"Type", "Date", "Category", "Amount", "Note"}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("error writing CSV header: %w", err)
	}

	// Write income records
	for _, income := range incomes {
		record := []string{
			"Income",
			income.Date.Format("2006-01-02"),
			income.Category,
			strconv.FormatFloat(income.Amount, 'f', 2, 64),
			income.Note,
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("error writing income record to CSV: %w", err)
		}
	}

	// Write expense records
	for _, expense := range expenses {
		record := []string{
			"Expense",
			expense.Date.Format("2006-01-02"),
			expense.Category,
			strconv.FormatFloat(expense.Amount, 'f', 2, 64),
			expense.Note,
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("error writing expense record to CSV: %w", err)
		}
	}

	writer.Flush() // Ensure all data is written to buffer
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("csv writer error: %w", err)
	}

	return b.String(), nil
}

// GenerateTransactionsPDF generates a PDF report of transactions.
func (s *ReportService) GenerateTransactionsPDF(startDate, endDate time.Time) (*bytes.Buffer, error) {
	if s.incomeService == nil || s.expenseService == nil {
		return nil, fmt.Errorf("report service is not properly initialized with income/expense services")
	}

	incomes, err := s.incomeService.GetIncomesByDateRange(startDate, endDate) // UserID removed
	if err != nil {
		return nil, fmt.Errorf("error fetching income data for PDF: %w", err)
	}

	expenses, err := s.expenseService.GetExpensesByDateRange(startDate, endDate) // UserID removed
	if err != nil {
		return nil, fmt.Errorf("error fetching expense data for PDF: %w", err)
	}

	// Calculate summary directly
	var totalIncome float64
	for _, item := range incomes {
		totalIncome += item.Amount
	}
	var totalExpenses float64
	for _, item := range expenses {
		totalExpenses += item.Amount
	}
	netBalance := totalIncome - totalExpenses

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// Title
	pdf.Cell(40, 10, "Financial Report")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	// pdf.Cell(40, 10, fmt.Sprintf("User ID: %d", userID)) // UserID removed from PDF content
	// pdf.Ln(5)
	pdf.Cell(40, 10, fmt.Sprintf("Period: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")))
	pdf.Ln(10)

	// Summary Section
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Summary")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 8, fmt.Sprintf("Total Income: %.2f", totalIncome))
	pdf.Ln(6)
	pdf.Cell(40, 8, fmt.Sprintf("Total Expenses: %.2f", totalExpenses))
	pdf.Ln(6)
	pdf.Cell(40, 8, fmt.Sprintf("Net Balance: %.2f", netBalance))
	pdf.Ln(10)

	// Table rendering helper
	renderTable := func(title string, headers []string, data [][]string) {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, title)
		pdf.Ln(8)

		pdf.SetFont("Arial", "B", 10)
		// Calculate column widths (simple equal distribution for now)
		// A4 width is 210mm. Margins usually 10mm each side. Usable width ~190mm.
		numCols := len(headers)
		colWidth := 190.0 / float64(numCols)

		// Headers
		pdf.SetFillColor(200, 200, 200) // Light grey for header
		for _, h := range headers {
			pdf.CellFormat(colWidth, 7, h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1) // Reset Y position to start of cell row

		// Data
		pdf.SetFont("Arial", "", 9)
		pdf.SetFillColor(230, 230, 230) // For alternating rows
		fill := false
		for _, row := range data {
			for i, item := range row {
				// Handle potential overflow for 'Note' column if it's too long
				// For simplicity, we're not doing MultiCell here yet.
				// A more robust solution would check content length and use MultiCell or truncate.
				if headers[i] == "Note" && len(item) > int(colWidth/2) { // Heuristic for when to truncate
					item = item[:int(colWidth/2)-3] + "..."
				}
				pdf.CellFormat(colWidth, 6, item, "1", 0, "L", fill, 0, "")
			}
			pdf.Ln(-1)
			fill = !fill
		}
		pdf.Ln(10)
	}

	// Income Transactions
	incomeHeaders := []string{"Date", "Category", "Amount", "Note"}
	var incomeData [][]string
	for _, item := range incomes {
		incomeData = append(incomeData, []string{
			item.Date.Format("2006-01-02"),
			item.Category,
			strconv.FormatFloat(item.Amount, 'f', 2, 64),
			item.Note,
		})
	}
	renderTable("Income Transactions", incomeHeaders, incomeData)

	// Expense Transactions
	expenseHeaders := []string{"Date", "Category", "Amount", "Note"}
	var expenseData [][]string
	for _, item := range expenses {
		expenseData = append(expenseData, []string{
			item.Date.Format("2006-01-02"),
			item.Category,
			strconv.FormatFloat(item.Amount, 'f', 2, 64),
			item.Note,
		})
	}
	renderTable("Expense Transactions", expenseHeaders, expenseData)

	// Footer (example for page number)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
		pdf.Ln(4)
		pdf.CellFormat(0, 10, "Generated: "+time.Now().Format("2006-01-02 15:04:05"), "", 0, "C", false, 0, "")
	})
	pdf.AliasNbPages("{nb}") // Define alias for total page numbers

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		log.Printf("Error generating PDF output: %v", err)
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	return &buffer, nil
}
