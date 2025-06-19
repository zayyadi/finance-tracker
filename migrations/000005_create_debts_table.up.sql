CREATE TABLE IF NOT EXISTS Debts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES Users(id) ON DELETE CASCADE,
    debtor_name VARCHAR(255) NOT NULL,
    description TEXT, -- Added for more details
    amount DECIMAL(10, 2) NOT NULL,
    due_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'Pending', -- e.g., Pending, Paid, Overdue
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_debts_user_id ON Debts(user_id);
CREATE INDEX IF NOT EXISTS idx_debts_due_date ON Debts(due_date);
