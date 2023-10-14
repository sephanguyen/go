CREATE TABLE public.billing_schedule_period (
    id integer NOT NULL,
    name text NOT NULL,
    billing_schedule_id integer NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    billing_date timestamp with time zone NOT NULL,
    remarks text NULL,
    is_archived boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.billing_schedule_period
    ADD CONSTRAINT billing_schedule_period_pk PRIMARY KEY (id);

CREATE SEQUENCE public.billing_schedule_period_id_seq
    AS integer;

ALTER SEQUENCE public.billing_schedule_period_id_seq OWNED BY public.billing_schedule_period.id;

ALTER TABLE ONLY public.billing_schedule_period ALTER COLUMN id SET DEFAULT nextval('public.billing_schedule_period_id_seq'::regclass);

CREATE POLICY rls_billing_schedule_period ON "billing_schedule_period" USING (permission_check(resource_path, 'billing_schedule_period')) WITH CHECK (permission_check(resource_path, 'billing_schedule_period'));

ALTER TABLE "billing_schedule_period" ENABLE ROW LEVEL security;
ALTER TABLE "billing_schedule_period" FORCE ROW LEVEL security;
