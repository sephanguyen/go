CREATE TABLE public.user_discount_tag (
    user_id text NOT NULL,
    location_id text NOT NULL,
    product_id text NOT NULL,
    discount_type text NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE public.user_discount_tag
    ADD CONSTRAINT user_discount_tag_pk PRIMARY KEY (user_id, location_id, product_id, discount_type);

ALTER TABLE public.user_discount_tag
    ADD CONSTRAINT fk_user_discount_tag_user_id FOREIGN KEY (user_id) REFERENCES public.users(user_id);

ALTER TABLE public.user_discount_tag
    ADD CONSTRAINT fk_user_discount_tag_location_id FOREIGN KEY (location_id) REFERENCES public.locations(location_id);

ALTER TABLE public.user_discount_tag
    ADD CONSTRAINT fk_user_discount_tag_product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id);

CREATE POLICY rls_user_discount_tag ON "user_discount_tag"
    using (permission_check(resource_path, 'user_discount_tag'))
    with check (permission_check(resource_path, 'user_discount_tag'));

CREATE POLICY rls_user_discount_tag_restrictive ON "user_discount_tag"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'user_discount_tag'))
    WITH CHECK (permission_check(resource_path, 'user_discount_tag'));

ALTER TABLE "user_discount_tag" ENABLE ROW LEVEL security;
ALTER TABLE "user_discount_tag" FORCE ROW LEVEL security;
