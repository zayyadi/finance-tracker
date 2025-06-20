package models

// CategoryExpenseStat represents the total expenses for a category.
type CategoryExpenseStat struct {
	Category    string  `json:"category"`
	TotalAmount float64 `json:"total_amount"`
}

// MonthlyTrendStat represents the income and expenses for a month.
type MonthlyTrendStat struct {
	Month         string  `json:"month"` // Format: YYYY-MM
	TotalIncome   float64 `json:"total_income"`
	TotalExpenses float64 `json:"total_expenses"`
}
