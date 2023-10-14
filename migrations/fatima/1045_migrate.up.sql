ALTER TABLE ONLY public.product DROP COLUMN billing_ratio_type_id;
ALTER TABLE ONLY public.billing_ratio DROP COLUMN billing_ratio_type_id;
DROP TABLE public.billing_ratio_type;
