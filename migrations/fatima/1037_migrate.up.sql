ALTER TABLE bill_item ADD COLUMN discount_amount numeric(12,2) NOT NULL;
ALTER TABLE bill_item ADD COLUMN tax_amount numeric(12,2) NOT NULL;
ALTER TABLE bill_item ADD COLUMN final_price numeric(12,2) NOT NULL;
