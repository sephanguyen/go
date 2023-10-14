ALTER TABLE public.bill_item
  DROP COLUMN IF EXISTS product_id;

ALTER TABLE public.bill_item ALTER COLUMN discount_amount_value DROP NOT NULL;
ALTER TABLE public.bill_item ALTER COLUMN discount_amount DROP NOT NULL;
ALTER TABLE public.bill_item ALTER COLUMN product_pricing DROP NOT NULL;

ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_location_id FOREIGN KEY (location_id) REFERENCES public.locations(location_id);