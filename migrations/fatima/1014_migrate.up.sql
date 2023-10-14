CREATE TABLE public.tax (
    id integer NOT NULL,
    name text NOT NULL,
    tax_percentage integer NOT NULL,
    tax_category text NOT NULL,
    default_flag boolean DEFAULT false NOT NULL,
    is_archived boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.tax
    ADD CONSTRAINT tax_pk PRIMARY KEY (id);

ALTER TABLE ONLY public.tax
    ADD CONSTRAINT check_types 
    CHECK (tax_category = 'TAX_CATEGORY_NONE' OR tax_category = 'TAX_CATEGORY_INCLUSIVE' OR tax_category = 'TAX_CATEGORY_EXCLUSIVE');

CREATE UNIQUE INDEX on public.tax (default_flag) 
    where default_flag = true;

CREATE SEQUENCE public.tax_id_seq
    AS integer;
    
ALTER SEQUENCE public.tax_id_seq OWNED BY public.tax.id;

ALTER TABLE ONLY public.tax ALTER COLUMN id SET DEFAULT nextval('public.tax_id_seq'::regclass);

CREATE POLICY rls_tax ON "tax" using (permission_check(resource_path, 'tax')) with check (permission_check(resource_path, 'tax'));

ALTER TABLE "tax" ENABLE ROW LEVEL security;
ALTER TABLE "tax" FORCE ROW LEVEL security;
