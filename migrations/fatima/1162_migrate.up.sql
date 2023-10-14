ALTER TABLE public.user_discount_tag ADD COLUMN product_group_id text;

ALTER TABLE public.user_discount_tag
    ADD CONSTRAINT fk_user_discount_tag_product_group_id FOREIGN KEY (product_group_id) REFERENCES public.product_group(product_group_id);
