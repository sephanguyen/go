ALTER TABLE public.bill_item ALTER COLUMN tax_id DROP NOT NULL;
ALTER TABLE public.bill_item ALTER COLUMN tax_category DROP NOT NULL;
ALTER TABLE public.bill_item ALTER COLUMN tax_percentage DROP NOT NULL;
ALTER TABLE public.bill_item ALTER COLUMN tax_amount DROP NOT NULL;
