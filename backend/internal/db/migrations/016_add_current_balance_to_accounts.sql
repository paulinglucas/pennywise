-- Track current balance for debt accounts synced via SimpleFIN
ALTER TABLE accounts ADD COLUMN current_balance REAL;
