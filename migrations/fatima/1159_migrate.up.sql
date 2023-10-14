CREATE TABLE IF NOT EXISTS public.order_leaving_reason (
    order_id text NOT NULL,
    leaving_reason_id text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT order_leaving_reason_pk PRIMARY KEY (order_id, leaving_reason_id),
    CONSTRAINT fk_order_leaving_reason_order_id FOREIGN KEY (order_id) REFERENCES public.order(order_id),
    CONSTRAINT fk_order_leaving_reason_leaving_reason_id FOREIGN KEY (leaving_reason_id) REFERENCES public.leaving_reason(leaving_reason_id)
);

CREATE POLICY rls_order_leaving_reason ON "order_leaving_reason"
    USING (permission_check(resource_path, 'order_leaving_reason'))
    WITH CHECK (permission_check(resource_path, 'order_leaving_reason'));

CREATE POLICY rls_order_leaving_reason_restrictive ON "order_leaving_reason"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'order_leaving_reason'))
    WITH CHECK (permission_check(resource_path, 'order_leaving_reason'));

ALTER TABLE "order_leaving_reason" ENABLE ROW LEVEL security;
ALTER TABLE "order_leaving_reason" FORCE ROW LEVEL security;
ALTER TABLE public.order DROP COLUMN IF EXISTS reason;
