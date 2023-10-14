CREATE TABLE public.billing_ratio (
    id integer NOT NULL,
    billing_ratio_type_id integer NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    billing_schedule_period_id integer NOT NULL,
    billing_ratio_numerator integer NOT NULL,
    billing_ratio_denominator integer NOT NULL,
    is_archived boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.billing_ratio
    ADD CONSTRAINT billing_ratio_pk PRIMARY KEY (id);

CREATE SEQUENCE public.billing_ratio_id_seq
    AS integer;
    
ALTER SEQUENCE public.billing_ratio_id_seq OWNED BY public.billing_ratio.id;

ALTER TABLE ONLY public.billing_ratio ALTER COLUMN id SET DEFAULT nextval('public.billing_ratio_id_seq'::regclass);

ALTER TABLE public.billing_ratio ADD CONSTRAINT fk_billing_ratio_type_id FOREIGN KEY(billing_ratio_type_id) REFERENCES billing_ratio_type(id);
ALTER TABLE public.billing_ratio ADD CONSTRAINT fk_billing_schedule_period_id FOREIGN KEY(billing_schedule_period_id) REFERENCES billing_schedule_period(id);

CREATE POLICY rls_billing_ratio ON "billing_ratio" using (permission_check(resource_path, 'billing_ratio')) with check (permission_check(resource_path, 'billing_ratio'));

ALTER TABLE "billing_ratio" ENABLE ROW LEVEL security;
ALTER TABLE "billing_ratio" FORCE ROW LEVEL security;
