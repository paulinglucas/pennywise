-- Link transactions to groups for split transaction support
ALTER TABLE transactions ADD COLUMN group_id TEXT REFERENCES transaction_groups(id);

CREATE INDEX IF NOT EXISTS idx_transactions_group ON transactions(group_id);
