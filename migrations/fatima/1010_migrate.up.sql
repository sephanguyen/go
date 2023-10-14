CREATE TABLE public.discount (
    id integer NOT NULL,
    name text NOT NULL,
    discount_type text NOT NULL,
    discount_amount_type text NOT NULL,
    discount_amount_value numeric(12,2) NOT NULL,
    recurring_valid_duration integer NULL,
    available_from timestamp with time zone NOT NULL,
    available_until timestamp with time zone NOT NULL,
    remarks text NULL,
    is_archived boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.discount
    ADD CONSTRAINT discount_pk PRIMARY KEY (id);

CREATE SEQUENCE public.discount_id_seq
    AS integer;
    
ALTER SEQUENCE public.discount_id_seq OWNED BY public.discount.id;

ALTER TABLE ONLY public.discount ALTER COLUMN id SET DEFAULT nextval('public.discount_id_seq'::regclass);

CREATE POLICY rls_discount ON "discount" using (permission_check(resource_path, 'discount')) with check (permission_check(resource_path, 'discount'));

ALTER TABLE "discount" ENABLE ROW LEVEL security;
ALTER TABLE "discount" FORCE ROW LEVEL security;