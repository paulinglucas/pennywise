-- mortgage-specific fields for debt accounts linked via SimpleFIN
ALTER TABLE accounts ADD COLUMN interest_rate REAL;
ALTER TABLE accounts ADD COLUMN loan_term_months INTEGER;
ALTER TABLE accounts ADD COLUMN purchase_price REAL;
ALTER TABLE accounts ADD COLUMN purchase_date TEXT;
ALTER TABLE accounts ADD COLUMN down_payment_pct REAL;
