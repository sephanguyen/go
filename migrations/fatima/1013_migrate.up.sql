CREATE TABLE public.product (
    id integer NOT NULL,
    name text NOT NULL,
    product_type text NOT NULL,
    tax_id integer,
    available_from timestamp with time zone NOT NULL,
    available_util timestamp with time zone NOT NULL,
    remarks text NULL,
    custom_billing_period timestamp with time zone,
    billing_schedule_id integer,
    disable_pro_rating_flag boolean DEFAULT false NOT NULL,
    is_archived boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.product
    ADD CONSTRAINT product_pk PRIMARY KEY (id);

CREATE SEQUENCE public.product_id_seq
    AS integer;

ALTER SEQUENCE public.product_id_seq OWNED BY public.product.id;

ALTER TABLE ONLY public.product ALTER COLUMN id SET DEFAULT nextval('public.product_id_seq'::regclass);

CREATE POLICY rls_product ON "product" USING (permission_check(resource_path, 'product')) WITH CHECK (permission_check(resource_path, 'product'));

ALTER TABLE "product" ENABLE ROW LEVEL security;
ALTER TABLE "product" FORCE ROW LEVEL security;
