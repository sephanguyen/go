CREATE TABLE IF NOT EXISTS public.discount_tag (
    discount_tag_id TEXT NOT NULL,
    discount_tag_name TEXT NOT NULL,
    selectable BOOLEAN DEFAULT TRUE NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    is_archived boolean DEFAULT false NOT NULL,
    CONSTRAINT pk__discount_tag PRIMARY KEY (discount_tag_id)
);

CREATE POLICY rls_discount_tag ON "discount_tag"
    USING (permission_check(resource_path, 'discount_tag'))
    with check (permission_check(resource_path, 'discount_tag'));

CREATE POLICY rls_discount_tag_restrictive ON "discount_tag" AS RESTRICTIVE TO PUBLIC
    USING (permission_check(resource_path, 'discount_tag'))
    with check (permission_check(resource_path, 'discount_tag'));

ALTER TABLE "discount_tag" ENABLE ROW LEVEL security;
ALTER TABLE "discount_tag" FORCE ROW LEVEL security;


ALTER TABLE public.user_discount_tag ADD COLUMN IF NOT EXISTS discount_tag_id TEXT NOT NULL;

ALTER TABLE public.user_discount_tag
    ADD CONSTRAINT fk_user_discount_tag_discount_tag_id FOREIGN KEY (discount_tag_id) REFERENCES public.discount_tag(discount_tag_id);
