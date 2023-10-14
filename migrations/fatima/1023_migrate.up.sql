ALTER TABLE public.product ADD COLUMN billing_ratio_type_id integer;
ALTER TABLE public.product ADD CONSTRAINT fk_billing_ratio_type_id FOREIGN KEY(billing_ratio_type_id) REFERENCES billing_ratio_type(id);
