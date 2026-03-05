-- Create transactions and transaction_tags tables
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    account_id TEXT NOT NULL REFERENCES accounts(id),
    type TEXT NOT NULL,
    category TEXT NOT NULL,
    amount REAL NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    date DATE NOT NULL,
    notes TEXT,
    is_recurring INTEGER NOT NULL DEFAULT 0,
    recurring_transaction_id TEXT REFERENCES recurring_transactions(id),
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions(user_id, date);
CREATE INDEX IF NOT EXISTS idx_transactions_user_category ON transactions(user_id, category);
CREATE INDEX IF NOT EXISTS idx_transactions_user_account ON transactions(user_id, account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_deleted ON transactions(deleted_at);

CREATE TABLE IF NOT EXISTS transaction_tags (
    id TEXT PRIMARY KEY,
    transaction_id TEXT NOT NULL REFERENCES transactions(id),
    tag TEXT NOT NULL,
    UNIQUE(transaction_id, tag)
);
