CREATE TABLE IF NOT EXISTS FinancialSummaries (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES Users(id) ON DELETE CASCADE,
    summary_type VARCHAR(10) NOT NULL, -- 'weekly', 'monthly', 'yearly'
    period_start_date DATE NOT NULL,
    period_end_date DATE NOT NULL,
    total_income DECIMAL(12, 2) NOT NULL,
    total_expenses DECIMAL(12, 2) NOT NULL,
    net_balance DECIMAL(12, 2) NOT NULL, -- Changed from 'balance' for clarity
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_financial_summary UNIQUE (user_id, summary_type, period_start_date) -- Ensure unique summaries
);
CREATE INDEX IF NOT EXISTS idx_fs_user_id_type ON FinancialSummaries(user_id, summary_type);
