ALTER TABLE public.bill_item ADD COLUMN discount_id integer;
ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_discount_id FOREIGN KEY(discount_id) REFERENCES public.discount(discount_id);
