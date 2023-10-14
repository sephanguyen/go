ALTER TABLE public.user_discount_tag
    DROP CONSTRAINT IF EXISTS user_discount_tag_pk;

ALTER TABLE public.user_discount_tag
    ADD CONSTRAINT user_discount_tag_pk PRIMARY KEY (user_id, discount_type, created_at, resource_path);