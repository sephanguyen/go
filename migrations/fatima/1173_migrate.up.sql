ALTER TABLE ONLY public.product_group ALTER COLUMN group_tag DROP NOT NULL;
ALTER TABLE public.product_group ALTER COLUMN discount_type SET NOT NULL;
