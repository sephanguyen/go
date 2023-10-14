ALTER TABLE public.bill_item ADD COLUMN previous_bill_item_sequence_number int NULL;
ALTER TABLE public.bill_item ADD COLUMN previous_bill_item_status text NULL;
ALTER TABLE public.bill_item ADD COLUMN adjustment_price numeric(12,2) NULL;
ALTER TABLE public.bill_item ADD COLUMN is_latest_bill_item boolean NULL;
ALTER TABLE public.bill_item ADD COLUMN price numeric(12,2) NULL;
ALTER TABLE public.bill_item ADD COLUMN old_price numeric(12,2) NULL;
ALTER TABLE public.bill_item ADD COLUMN billing_ratio_numerator int NULL;
ALTER TABLE public.bill_item ADD COLUMN billing_ratio_denominator int NULL;

ALTER TABLE ONLY public.bill_item
    ADD CONSTRAINT fk_bill_item_location_id FOREIGN KEY(location_id) REFERENCES public.locations(location_id);
