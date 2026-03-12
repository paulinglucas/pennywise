-- Track SimpleFIN transaction IDs for deduplication during sync
ALTER TABLE transactions ADD COLUMN external_id TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_account_external_id
  ON transactions(account_id, external_id) WHERE external_id IS NOT NULL;
