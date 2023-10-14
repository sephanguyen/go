ALTER TABLE ONLY public.bill_item
    ADD COLUMN IF NOT EXISTS product_id TEXT,
    ADD COLUMN IF NOT EXISTS student_product_id TEXT,
    ADD COLUMN IF NOT EXISTS previous_bill_item_sequence_number INTEGER,
    ADD COLUMN IF NOT EXISTS previous_bill_item_status TEXT,
    ADD COLUMN IF NOT EXISTS adjustment_price NUMERIC(12,2),
    ADD COLUMN IF NOT EXISTS is_latest_bill_item BOOLEAN,
    ADD COLUMN IF NOT EXISTS price NUMERIC(12,2),
    ADD COLUMN IF NOT EXISTS old_price NUMERIC(12,2),
    ADD COLUMN IF NOT EXISTS billing_ratio_numerator INTEGER,
    ADD COLUMN IF NOT EXISTS billing_ratio_denominator INTEGER,
    ADD COLUMN IF NOT EXISTS is_reviewed BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS raw_discount_amount NUMERIC(12,2);