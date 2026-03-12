-- Track debt account balance over time for accurate historical net worth
CREATE TABLE IF NOT EXISTS account_balance_history (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(id),
    balance REAL NOT NULL,
    recorded_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_account_balance_history_account_recorded ON account_balance_history(account_id, recorded_at);
